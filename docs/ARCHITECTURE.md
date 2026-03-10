# Arquitetura da Plataforma de Processamento de Vídeo

## Sumário

1. [Visão Geral do Sistema](#1-visão-geral-do-sistema)
2. [Estilo Arquitetural](#2-estilo-arquitetural)
3. [Contexto do Sistema — C4 Nível 1](#3-contexto-do-sistema--c4-nível-1)
4. [Mapa de Serviços — C4 Nível 2](#4-mapa-de-serviços--c4-nível-2)
5. [Padrões de Comunicação](#5-padrões-de-comunicação)
6. [Fluxo de Processamento de Vídeo](#6-fluxo-de-processamento-de-vídeo)
7. [Fluxo de Autenticação](#7-fluxo-de-autenticação)
8. [Modelo de Dados](#8-modelo-de-dados)
9. [Bibliotecas Compartilhadas](#9-bibliotecas-compartilhadas)
10. [Infraestrutura e Monitoramento](#10-infraestrutura-e-monitoramento)
11. [Considerações de Escalabilidade](#11-considerações-de-escalabilidade)
12. [Arquitetura de Segurança](#12-arquitetura-de-segurança)

---

## 1. Visão Geral do Sistema

A **Plataforma de Processamento de Vídeo** é um sistema de microserviços orientado a eventos que permite que usuários façam upload de arquivos de vídeo, processem-nos automaticamente para extração de frames via FFmpeg e façam download dos frames como um arquivo ZIP. O sistema é projetado para ser escalável horizontalmente, resiliente a falhas e observável em produção.

**Principais capacidades:**

- Upload seguro de vídeos (até 500 MB, formatos `.mp4`, `.avi`, `.mov`, `.mkv`, `.webm`)
- Extração assíncrona de frames a 1 FPS via FFmpeg, executada por workers escaláveis
- Criação de arquivo ZIP com todos os frames extraídos
- Notificação por e-mail ao término do processamento (sucesso ou falha)
- Download dos frames via presigned URL do S3 (sem tráfego pelos serviços)
- Expiração e limpeza automática dos dados após 15 dias
- Observabilidade completa com Prometheus, Grafana e Alertmanager

**Stack tecnológica:**

| Camada | Tecnologia |
|---|---|
| Backend | Go 1.24+, Chi Router, Uber FX, GORM |
| Frontend | React 18, TypeScript, Vite, TanStack Router/Query, shadcn/ui, Bun |
| Banco de dados | PostgreSQL 15 (3 schemas) |
| Cache / Sessão | Redis 7 |
| Mensageria | RabbitMQ 3 (AMQP) |
| Armazenamento | AWS S3 (2 buckets) |
| Infraestrutura | Docker Compose, Nginx |
| Monitoramento | Prometheus, Grafana, Alertmanager |
| CI/CD | GitHub Actions, ghcr.io |

---

## 2. Estilo Arquitetural

O sistema adota dois padrões arquiteturais complementares: **Microserviços** no nível do sistema e **Clean Architecture (Hexagonal)** no nível de cada serviço.

### Microserviços

Cada serviço é um processo independente com banco de dados lógico próprio (schema separado no PostgreSQL), implantado em container Docker. Os serviços se comunicam via REST/HTTP para operações síncronas e via RabbitMQ para processamento assíncrono. Nenhum serviço acessa o schema do outro diretamente.

### Clean Architecture por Serviço

Todos os serviços Go seguem uma estrutura de camadas com dependências apontando sempre para dentro (ver [ADR-001](docs/adr/ADR-001-clean-architecture.md)):

```
┌─────────────────────────────────────┐
│           Infrastructure            │  ← GORM, S3, RabbitMQ, SMTP, Redis
├─────────────────────────────────────┤
│           Controllers               │  ← HTTP handlers, AMQP consumers
├─────────────────────────────────────┤
│           Use Cases                 │  ← Regras de negócio
├─────────────────────────────────────┤
│     Domain (Entities + Ports)       │  ← Interfaces e entidades puras
└─────────────────────────────────────┘
         Dependências ↑ apenas
```

O **Domain** não importa nenhum pacote externo. As **Use Cases** dependem apenas do Domain. A **Infrastructure** implementa as interfaces (ports) definidas no Domain. Esta estrutura garante que a lógica de negócio seja testável sem infraestrutura, resultando em cobertura média de **87,7%** de testes entre todos os serviços.

A injeção de dependência é gerenciada pelo **Uber FX** ([ADR-006](docs/adr/ADR-006-uber-fx-dependency-injection.md)), que valida o grafo completo de dependências na inicialização e orquestra o desligamento gracioso na ordem inversa.

---

## 3. Contexto do Sistema — C4 Nível 1

Visão de alto nível: quem usa o sistema e com quais sistemas externos ele se integra.

```mermaid
graph TD
    U(["👤 Usuário\n(navegador)"])

    subgraph Sistema["Plataforma de Processamento de Vídeo"]
        UI["React UI\nPorta 3000"]
        PLATFORM["5 Microserviços Go\n+ Infraestrutura"]
    end

    subgraph Externos["Sistemas Externos"]
        S3["☁️ AWS S3\n2 buckets"]
        EMAIL["📧 Gmail SMTP\nsmtp.gmail.com:587"]
    end

    U -->|"HTTPS — upload, visualização,\ndownload de frames"| UI
    UI -->|"REST/JSON"| PLATFORM
    PLATFORM -->|"S3 API — armazena vídeos,\nframes e ZIPs"| S3
    PLATFORM -->|"SMTP — notificações\nde conclusão/falha"| EMAIL
    S3 -->|"Download direto via\npresigned URL"| U
```

---

## 4. Mapa de Serviços — C4 Nível 2

Visão detalhada de todos os containers, suas responsabilidades, portas e dependências.

```mermaid
graph TD
    U(["👤 Usuário"])

    subgraph Frontend
        UI["React UI\n──────────\nPorta: 3000\nNginx + Vite + Bun\nTanStack Router/Query\nshadcn/ui + TailwindCSS"]
    end

    subgraph Services["Serviços Go"]
        GW["API Gateway\n──────────\nPorta: 8080\nChi Router\nUpload, listagem, download\nValidação de arquivo\nGeração de presigned URL"]
        AUTH["Auth Service\n──────────\nPorta: 8081\nRegistro e login\nJWT HS256 (15min/7d)\nbcrypt + refresh tokens"]
        STOR["Storage Service\n──────────\nPorta: 8082\nCriação de ZIP em memória\nAcesso interno apenas"]
        WORKER["Processing Workers\n──────────\nSem porta (background)\n3 réplicas\nFFmpeg (1 FPS)\nExtraçao e upload de frames"]
        NOTIF["Notification Service\n──────────\nSem porta (background)\nConsome fila de notificações\nEnvia e-mail via SMTP"]
        CRON["Cleanup Cron\n──────────\nDiário às 02:00 UTC\nSupercronic\nExpira vídeos após 15 dias"]
    end

    subgraph Infra["Infraestrutura"]
        PG[("PostgreSQL 15\nPorta: 5432\n3 schemas\nauth, videos,\nnotifications")]
        REDIS[("Redis 7\nPorta: 6379\nBlacklist JWT\nCache de sessão\nRate limiting")]
        MQ[("RabbitMQ 3\nPortas: 5672, 15672\nvideo.processing.queue\nvideo.notification.queue")]
        S3[("AWS S3\n2 buckets\nuploads / processed")]
    end

    subgraph Monitoring["Monitoramento"]
        PROM["Prometheus\nPorta: 9090"]
        GRAF["Grafana\nPorta: 3001\n4 dashboards"]
        ALERT["Alertmanager\nPorta: 9093\n10 regras"]
    end

    U -->|"HTTPS"| UI
    UI -->|"REST/JSON"| GW
    GW -->|"HTTP — valida JWT"| AUTH
    GW -->|"SQL"| PG
    GW -->|"Redis"| REDIS
    GW -->|"S3 API — upload/presigned URL"| S3
    GW -->|"AMQP — publica job"| MQ
    GW -->|"HTTP — cria ZIP"| STOR
    AUTH -->|"SQL"| PG
    AUTH -->|"Redis"| REDIS
    MQ -->|"AMQP — consome job"| WORKER
    WORKER -->|"SQL"| PG
    WORKER -->|"S3 API — download/upload"| S3
    WORKER -->|"HTTP — cria ZIP"| STOR
    WORKER -->|"AMQP — publica notificação"| MQ
    STOR -->|"S3 API"| S3
    MQ -->|"AMQP — consome notificação"| NOTIF
    NOTIF -->|"SQL"| PG
    CRON -->|"SQL"| PG
    CRON -->|"S3 API — deleta objetos expirados"| S3

    GW & AUTH & STOR & WORKER & NOTIF -->|"métricas /metrics"| PROM
    PROM --> GRAF
    PROM --> ALERT
```

---

## 5. Padrões de Comunicação

O sistema utiliza dois padrões de comunicação distintos, cada um aplicado conforme a natureza da operação.

```mermaid
flowchart LR
    subgraph Sincrono["Síncrono — REST/HTTP"]
        direction TB
        UI2["React UI"]
        GW2["API Gateway"]
        AUTH2["Auth Service\n(validação JWT)"]
        STOR2["Storage Service\n(criação de ZIP)"]
        DB2[("PostgreSQL\nRedis\nS3")]

        UI2 -->|"POST /videos/upload\nGET /videos\nGET /videos/{id}/status\nGET /videos/{id}/download"| GW2
        GW2 -->|"validação de token"| AUTH2
        GW2 -->|"POST /internal/zip/create"| STOR2
        GW2 --- DB2
    end

    subgraph Assincrono["Assíncrono — RabbitMQ AMQP"]
        direction TB
        GW3["API Gateway\n(produtor)"]
        Q1[["video.processing.queue\n(durable, persistent)"]]
        W1["Processing Worker × 3\n(QoS prefetch=1)"]
        Q2[["video.notification.queue\n(durable, persistent)"]]
        N1["Notification Service\n(consumidor)"]

        GW3 -->|"publica job após upload"| Q1
        Q1 -->|"distribui 1 job por worker"| W1
        W1 -->|"publica evento de conclusão"| Q2
        Q2 -->|"consome e envia e-mail"| N1
    end
```

**Regras de decisão:**

| Critério | REST/HTTP | RabbitMQ AMQP |
|---|---|---|
| Tempo de resposta | Imediato (necessário) | Deferível |
| Operação | Curta (< 1s) | Longa (processamento FFmpeg) |
| Acoplamento | Direto | Desacoplado |
| Exemplos | Login, upload (resposta 202), download | Processamento, notificação |

---

## 6. Fluxo de Processamento de Vídeo

Sequência completa desde o upload até o download do arquivo ZIP.

```mermaid
sequenceDiagram
    autonumber
    actor Usuario as Usuário
    participant UI as React UI
    participant GW as API Gateway
    participant PG as PostgreSQL
    participant S3up as S3 (uploads)
    participant MQ as RabbitMQ
    participant Worker as Processing Worker
    participant S3proc as S3 (processed)
    participant StorSvc as Storage Service
    participant Notif as Notification Service
    participant SMTP as Gmail SMTP

    Note over Usuario,SMTP: Fase 1 — Upload

    Usuario->>UI: seleciona arquivo de vídeo
    UI->>GW: POST /api/v1/videos/upload\n(Authorization: Bearer token)
    GW->>GW: valida JWT + tamanho (≤500MB)\n+ extensão (.mp4/.avi/.mov/.mkv/.webm)
    GW->>S3up: stream multipart upload\nuploads/{video_id}/{filename}
    S3up-->>GW: OK
    GW->>PG: INSERT videos (status=PENDING,\nexpires_at=NOW()+15d)
    GW->>MQ: publish video.processing.queue\n{video_id, user_id, s3_key, filename}
    GW-->>UI: 202 Accepted {video_id}
    UI-->>Usuario: "Upload recebido, processando..."

    Note over Worker,S3proc: Fase 2 — Processamento (assíncrono)

    MQ->>Worker: consume video.processing.queue\n(prefetch=1)
    Worker->>PG: UPDATE status=PROCESSING,\nstarted_at=NOW()
    Worker->>S3up: download video para /tmp/
    Worker->>Worker: ffmpeg -i video.mp4\n-vf fps=1 -qscale:v 2\nframe_%04d.jpg
    Worker->>S3proc: upload frames\nprocessed/{id}/frames/frame_*.jpg
    Worker->>StorSvc: POST /internal/zip/create\n{video_id, s3_prefix, output_key}
    StorSvc->>S3proc: lista todos os frames
    StorSvc->>StorSvc: cria ZIP em memória
    StorSvc->>S3proc: upload processed/{id}/{filename}.zip
    StorSvc-->>Worker: {zip_path, file_count, zip_size_bytes}
    Worker->>PG: UPDATE status=COMPLETED,\nframe_count, zip_path, completed_at
    Worker->>MQ: publish video.notification.queue\n{video_id, user_id, status, frame_count}
    Worker->>MQ: ack mensagem processada

    Note over Notif,SMTP: Fase 3 — Notificação (assíncrono)

    MQ->>Notif: consume video.notification.queue
    Notif->>PG: INSERT notification_log (PENDING)
    Notif->>SMTP: envia e-mail de conclusão\n"X frames extraídos — faça o download"
    Notif->>PG: UPDATE notification_log (SENT)

    Note over Usuario,S3proc: Fase 4 — Download

    Usuario->>UI: clica em "Download"
    UI->>GW: GET /api/v1/videos/{id}/download
    GW->>PG: verifica ownership + status=COMPLETED
    GW->>S3proc: GeneratePresignedURL\n(TTL=15 minutos)
    S3proc-->>GW: presigned URL
    GW-->>UI: {download_url, filename, expires_in: 900}
    UI-->>Usuario: redirect para presigned URL
    Usuario->>S3proc: download direto do ZIP\n(sem passar pelos serviços)
```

**Ciclo de vida do status no banco de dados:**

```mermaid
stateDiagram-v2
    [*] --> PENDING: upload recebido
    PENDING --> PROCESSING: worker consome job
    PROCESSING --> COMPLETED: FFmpeg + ZIP concluídos
    PROCESSING --> FAILED: erro FFmpeg/S3
    COMPLETED --> [*]: cleanup após 15 dias
    FAILED --> [*]: cleanup após 15 dias
```

---

## 7. Fluxo de Autenticação

Ciclo de vida completo de tokens JWT com blacklist no Redis.

```mermaid
sequenceDiagram
    autonumber
    actor Usuario as Usuário
    participant UI as React UI
    participant GW as API Gateway
    participant Auth as Auth Service
    participant PG as PostgreSQL
    participant Redis as Redis

    Note over Usuario,Redis: Registro

    Usuario->>UI: preenche formulário de registro
    UI->>Auth: POST /api/v1/auth/register\n{username, email, password}
    Auth->>Auth: bcrypt.GenerateFromPassword(password)
    Auth->>PG: INSERT auth.users (password_hash)
    Auth->>Auth: gera access_token (15min, HS256)\n+ refresh_token (7 dias)
    Auth->>PG: INSERT auth.refresh_tokens
    Auth-->>UI: {access_token, refresh_token}

    Note over Usuario,Redis: Login

    Usuario->>UI: preenche formulário de login
    UI->>Auth: POST /api/v1/auth/login\n{email, password}
    Auth->>PG: SELECT user BY email
    Auth->>Auth: bcrypt.CompareHashAndPassword()
    Auth->>Auth: gera novo par de tokens
    Auth->>PG: INSERT auth.refresh_tokens
    Auth-->>UI: {access_token, refresh_token}

    Note over Usuario,Redis: Requisição autenticada

    Usuario->>UI: acessa recurso protegido
    UI->>GW: GET /api/v1/videos\nAuthorization: Bearer {access_token}
    GW->>GW: middleware JWT:\n1. extrai Bearer token\n2. valida assinatura HS256\n3. verifica exp
    GW->>Redis: SISMEMBER jwt:blacklist {jti}
    Redis-->>GW: 0 (não está na blacklist)
    GW->>GW: injeta {user_id, username, email}\nno contexto da requisição
    GW-->>UI: 200 OK {dados}

    Note over Usuario,Redis: Logout

    Usuario->>UI: clica em "Sair"
    UI->>Auth: POST /api/v1/auth/logout\n{refresh_token}
    Auth->>PG: DELETE auth.refresh_tokens WHERE token=?
    Auth->>Redis: SETEX jwt:blacklist:{jti} 900\n(TTL = duração restante do access_token)
    Auth-->>UI: 200 OK

    Note over Usuario,Redis: Renovação do token

    UI->>Auth: POST /api/v1/auth/refresh\n{refresh_token}
    Auth->>PG: SELECT refresh_token WHERE token=? AND expires_at > NOW()
    Auth->>Auth: gera novo access_token (15min)
    Auth-->>UI: {access_token}
```

---

## 8. Modelo de Dados

O banco de dados `video_platform` é dividido em três schemas. Foreign keys cruzam schemas para manter integridade referencial com mínimo acoplamento entre serviços.

```mermaid
erDiagram
    auth_users {
        int id PK
        varchar username UK
        varchar email UK
        varchar password_hash
        timestamp created_at
        timestamp updated_at
    }

    auth_refresh_tokens {
        int id PK
        int user_id FK
        varchar token UK
        timestamp expires_at
        timestamp created_at
    }

    videos_videos {
        uuid id PK
        int user_id FK
        varchar filename
        text original_path
        varchar status
        int fps
        int frame_count
        text zip_path
        text error_message
        timestamp created_at
        timestamp started_at
        timestamp completed_at
        timestamp expires_at
    }

    notifications_log {
        int id PK
        int user_id FK
        uuid video_id FK
        varchar type
        varchar status
        text recipient
        text subject
        text error_message
        timestamp sent_at
        timestamp created_at
    }

    auth_users ||--o{ auth_refresh_tokens : "possui"
    auth_users ||--o{ videos_videos : "possui"
    auth_users ||--o{ notifications_log : "recebe"
    videos_videos |o--o{ notifications_log : "gera"
```

**Índices relevantes:**

| Tabela | Índice | Colunas | Tipo |
|---|---|---|---|
| `videos.videos` | `idx_videos_user_status` | `(user_id, status)` | Composto |
| `videos.videos` | `idx_videos_created_at` | `(created_at)` | Simples |
| `videos.videos` | `idx_videos_expires_at` | `(expires_at) WHERE status='COMPLETED'` | Parcial |
| `notifications.notification_log` | `idx_notifications_user_id` | `(user_id)` | Simples |
| `notifications.notification_log` | `idx_notifications_video_id` | `(video_id)` | Simples |
| `notifications.notification_log` | `idx_notifications_status` | `(status)` | Simples |

**Enumerações e restrições:**

- `videos.status`: `CHECK IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED')`
- `notifications.type`: `CHECK IN ('EMAIL', 'WEBHOOK')`
- `notifications.status`: `CHECK IN ('PENDING', 'SENT', 'FAILED')`
- `videos.expires_at`: `DEFAULT CURRENT_TIMESTAMP + INTERVAL '15 days'`
- `videos.fps`: `DEFAULT 1`
- `notifications.video_id`: `ON DELETE SET NULL` (preserva log mesmo se vídeo for deletado)
- `auth.refresh_tokens.user_id`: `ON DELETE CASCADE`
- `videos.videos.user_id`: `ON DELETE CASCADE`

---

## 9. Bibliotecas Compartilhadas

O módulo `shared/` é referenciado localmente por todos os serviços via Go Workspace (`go.work`), eliminando a necessidade de publicar versões ([ADR-005](docs/adr/ADR-005-go-workspace-monorepo.md)). Mudanças propagam imediatamente para todos os serviços.

| Pacote | Responsabilidade | Serviços |
|---|---|---|
| `shared/pkg/config` | Carrega 50+ variáveis de ambiente (DB, JWT, AWS, SMTP, Redis, RabbitMQ) | Todos |
| `shared/pkg/database/postgres` | Conexão GORM com pool (25 max open, 5 idle, lifetime 5min) | Auth, Gateway, Worker, Notification |
| `shared/pkg/database/redis` | Cliente Redis para blacklist JWT, cache e rate limiting | Auth, Gateway |
| `shared/pkg/messaging/rabbitmq` | Publisher e Consumer AMQP — JSON persistente, prefetch=1, ack/nack manual | Gateway, Worker, Notification |
| `shared/pkg/auth/jwt` | Geração e validação de tokens HS256 + middleware HTTP Chi | Auth, Gateway |
| `shared/pkg/storage/s3` | Upload, download, multipart, presigned URL, listagem e deleção de objetos | Gateway, Worker, Storage |
| `shared/pkg/httpclient` | Cliente HTTP com retry automático (3×, timeout 30s) | Gateway, Worker |
| `shared/pkg/rest` | Helpers de resposta HTTP padronizados `{data, message, error}` | Todos os serviços HTTP |
| `shared/pkg/logging` | Logging estruturado em JSON com tag do serviço | Todos |
| `shared/pkg/metrics` | Métricas Prometheus + middleware HTTP para coleta automática | Todos |

---

## 10. Infraestrutura e Monitoramento

### Stack de Observabilidade

```mermaid
graph LR
    subgraph Servicos["Serviços (métricas /metrics)"]
        S1["Auth :8080/metrics"]
        S2["API Gateway :8080/metrics"]
        S3["Worker :8080/metrics"]
        S4["Storage :8080/metrics"]
        S5["Notification :8080/metrics"]
        S6["postgres-exporter :9187"]
        S7["redis-exporter :9121"]
        S8["rabbitmq :15692/metrics"]
    end

    PROM["Prometheus\n:9090\nscrape: 15s\nretention: 15d"]
    ALERT["Alertmanager\n:9093\nrouting por severidade"]
    GRAF["Grafana\n:3001\n4 dashboards"]
    EMAIL2["E-mail\n(alerta crítico/warning)"]

    Servicos -->|"pull /metrics"| PROM
    PROM -->|"avalia regras"| ALERT
    PROM -->|"datasource"| GRAF
    ALERT -->|"SMTP"| EMAIL2
```

### Dashboards Grafana

| Dashboard | Painéis Principais |
|---|---|
| **System Overview** | Taxa de requisições HTTP por serviço, taxa de erros 5xx, latência P95, requests in-flight, status up/down |
| **Video Processing** | Profundidade da fila RabbitMQ, taxa de processamento (sucesso/falha), duração P95 por operação, vídeos por status, taxa de frames extraídos, taxa de e-mails enviados |
| **Database Health** | Taxa de queries por serviço/operação, duração P95 de queries, conexões abertas, conexões ativas no PostgreSQL, uso de memória Redis, taxa de comandos Redis |
| **Infrastructure** | CPU por instância, memória por instância, uso de disco, I/O de rede (RX/TX), CPU de container, memória de container |

### Regras de Alerta

| Alerta | Condição | Duração | Severidade |
|---|---|---|---|
| `ServiceDown` | Qualquer serviço monitorado com `up == 0` | 2 min | **Critical** |
| `DiskSpaceLow` | Espaço livre em disco `< 10%` | 5 min | **Critical** |
| `HighHTTPErrorRate` | Taxa de respostas 5xx `> 5%` | 5 min | Warning |
| `HighRequestLatency` | Latência P95 `> 2s` | 5 min | Warning |
| `HighQueueDepth` | Mensagens pendentes na fila `> 1.000` | 10 min | Warning |
| `DatabaseConnectionPoolNearlyExhausted` | Pool de conexões `> 80%` usado | 5 min | Warning |
| `HighVideoProcessingFailureRate` | Taxa de falhas no processamento `> 10%` | 10 min | Warning |
| `HighEmailFailureRate` | Taxa de falhas no envio de e-mail `> 20%` | 5 min | Warning |
| `HighCPUUsage` | Uso de CPU `> 80%` | 10 min | Warning |
| `HighMemoryUsage` | Uso de memória `> 85%` | 5 min | Warning |

Alertas críticos são enviados imediatamente com prefixo `[CRITICAL]`; warnings são agrupados com `group_wait: 10s` e repetidos a cada 12h.

### Cleanup Cron

O job de limpeza é executado diariamente às **02:00 UTC** via [supercronic](https://github.com/aptible/supercronic) (cron seguro para containers):

1. Consulta `videos.videos WHERE expires_at < NOW()`
2. Para cada vídeo expirado: deleta o arquivo original do bucket `uploads`, os frames e o ZIP do bucket `processed`
3. Deleta o registro do PostgreSQL (cascade deleta os logs de notificação)
4. Registra contadores: vídeos deletados, objetos S3 removidos, duração da execução
5. Suporta modo `--dry-run` para validação sem deleção real

---

## 11. Considerações de Escalabilidade

### Escalabilidade Horizontal

```mermaid
graph TD
    LB["Load Balancer\n(Nginx / AWS ALB)"]

    subgraph Stateless["Serviços Stateless — escaláveis horizontalmente"]
        GW1["API Gateway\nreplica 1"]
        GW2["API Gateway\nreplica 2"]
        AUTH1["Auth Service\nreplica 1"]
        AUTH2["Auth Service\nreplica 2"]
        W1["Processing Worker\nreplica 1"]
        W2["Processing Worker\nreplica 2"]
        W3["Processing Worker\nreplica 3"]
    end

    subgraph SharedState["Estado Compartilhado — externo aos serviços"]
        PG2[("PostgreSQL\n(read replicas possíveis)")]
        REDIS2[("Redis\n(cluster)")]
        MQ2[("RabbitMQ\n(cluster)")]
        S32[("AWS S3\n(elasticidade nativa)")]
    end

    LB --> GW1 & GW2
    LB --> AUTH1 & AUTH2
    MQ2 --> W1 & W2 & W3

    GW1 & GW2 & AUTH1 & AUTH2 & W1 & W2 & W3 --- SharedState
```

**Por que os serviços são stateless:**

- Nenhum serviço armazena estado em memória entre requisições
- Sessões são representadas por JWT (validados localmente) + Redis (blacklist)
- Jobs são distribuídos pelo RabbitMQ com prefetch=1 (um job por worker por vez)
- Arquivos são armazenados no S3, não em disco local

**Gargalos e estratégias:**

| Componente | Estratégia de Escala |
|---|---|
| Processing Worker | Aumentar réplicas (Docker: `deploy.replicas`, K8s: HPA por profundidade da fila) |
| API Gateway | Réplicas stateless atrás de load balancer |
| PostgreSQL | Read replicas para queries de leitura; PgBouncer para pooling |
| S3 | Elasticidade automática da AWS |
| RabbitMQ | Clustering para alta disponibilidade |

---

## 12. Arquitetura de Segurança

O sistema implementa defesa em profundidade com múltiplas camadas de segurança.

```mermaid
graph TD
    U2(["👤 Usuário"])

    subgraph Perimetro["Perímetro"]
        RL["Rate Limiting\n10 uploads/hora/usuário\n(Redis)"]
        TLS["TLS/HTTPS\n(Nginx termina TLS)"]
    end

    subgraph Auth2["Autenticação e Autorização"]
        JWT2["JWT HS256\naccess: 15min\nrefresh: 7 dias"]
        BL["Blacklist de tokens\n(Redis, TTL=15min após logout)"]
        BCRYPT["Senhas com bcrypt\n(custo padrão)"]
    end

    subgraph Upload["Validação de Upload"]
        SIZE["Tamanho máximo: 500 MB"]
        EXT["Whitelist de extensões\n.mp4 .avi .mov .mkv .webm"]
    end

    subgraph Storage2["Armazenamento Seguro"]
        S3PRIV["Buckets S3 privados\n(sem acesso público)"]
        PRESIGN["Presigned URLs\nTTL: 15 minutos"]
        EXPIRE["Expiração automática\n15 dias"]
    end

    subgraph CICD["CI/CD — Análise de Segurança"]
        TRIVY["Trivy\n(vulnerabilidades em containers)"]
        GOSEC["gosec\n(análise estática Go)"]
        CODEQL["CodeQL\n(GitHub Advanced Security)"]
    end

    U2 --> TLS --> RL --> JWT2
    JWT2 --> BL
    JWT2 --> Upload
    Upload --> S3PRIV
    S3PRIV --> PRESIGN
```

**Resumo das camadas de defesa:**

| Ameaça | Controle |
|---|---|
| Senhas fracas / vazamento | `bcrypt` com custo padrão; apenas o hash é armazenado |
| Tokens roubados | Access token expira em 15min; logout invalida imediatamente via Redis blacklist |
| Abuso de upload | Rate limiting (10/hora/usuário via Redis); limite de 500 MB; whitelist de extensões |
| Upload de arquivos maliciosos | Whitelist de extensões (.mp4, .avi, .mov, .mkv, .webm) |
| Exposição de credenciais AWS | Downloads via presigned URL (TTL 15min); buckets S3 totalmente privados |
| Dados retidos indefinidamente | Expiração automática após 15 dias + cleanup cron diário |
| Vulnerabilidades em dependências | Trivy (containers) + gosec (Go) + CodeQL em cada PR |

---

*Decisões arquiteturais detalhadas estão documentadas nos [ADRs](docs/adr/).*
