# ADR-007: PostgreSQL com múltiplos schemas em vez de bancos separados por serviço

**Status:** Aceito
**Data:** 2026-02-01

---

## Contexto

A plataforma tem três serviços com necessidades de persistência: Auth Service, API Gateway e Notification Service. Em uma arquitetura de microsserviços estrita, cada serviço teria seu próprio banco de dados isolado (o padrão "database per service"). Na prática, isso cria desafios operacionais e de integridade referencial que precisam ser avaliados contra os benefícios do isolamento.

**Problema central de integridade:** a tabela `notifications.notification_log` precisa referenciar `auth.users` (para saber quem notificar) e `videos.videos` (para identificar o vídeo). Com bancos separados, essas chaves estrangeiras seriam impossíveis de enforçar no banco — a integridade ficaria a cargo da aplicação.

Três abordagens foram consideradas:

1. **Bancos de dados separados por serviço**: máximo isolamento; cada serviço gerencia seu próprio banco. Exige replicação eventual de dados (ex.: a tabela de usuários precisaria ser replicada para o serviço de notificação) ou APIs de lookup entre serviços para consistência.
2. **Um único banco com múltiplos schemas**: isolamento lógico por schema (`auth`, `videos`, `notifications`); chaves estrangeiras cross-schema mantêm integridade referencial; uma única instância PostgreSQL a gerenciar.
3. **Um único schema com todas as tabelas**: simples, mas sem separação de responsabilidades. Permissions e ownership confundidos.

---

## Decisão

Adotar um **único banco de dados PostgreSQL** (`video_platform`) com **três schemas distintos**, um por domínio funcional:

```sql
-- Schema de autenticação (gerenciado pelo Auth Service)
CREATE SCHEMA auth;
auth.users           -- usuários registrados
auth.refresh_tokens  -- tokens de refresh ativos

-- Schema de vídeos (gerenciado pelo API Gateway e Processing Worker)
CREATE SCHEMA videos;
videos.videos        -- registro de cada vídeo enviado

-- Schema de notificações (gerenciado pelo Notification Service)
CREATE SCHEMA notifications;
notifications.notification_log  -- log de notificações enviadas
```

**Chaves estrangeiras cross-schema** são usadas para manter integridade referencial:
```sql
videos.videos.user_id → auth.users.id (ON DELETE CASCADE)
notifications.notification_log.user_id → auth.users.id (ON DELETE CASCADE)
notifications.notification_log.video_id → videos.videos.id (ON DELETE SET NULL)
```

Todos os serviços se conectam ao mesmo banco usando o usuário `videoadmin`, com permissões concedidas explicitamente por schema. O isolamento é **lógico** (por schema e naming convention) em vez de físico (banco separado).

---

## Consequências

**Positivas:**
- Chaves estrangeiras cross-schema garantem integridade referencial no banco, sem necessidade de sincronização eventual ou APIs de lookup entre serviços.
- Uma única instância PostgreSQL a monitorar, fazer backup e escalar.
- Um único pool de conexões compartilhado — mais eficiente do que múltiplos pools para bancos separados em um ambiente com poucos serviços.
- Migrations mais simples: um único script `01_init.sql` cria todos os schemas, tabelas, índices e triggers.
- Queries de relatório ou debug que precisem cruzar dados de múltiplos domínios são possíveis diretamente no banco.

**Negativas:**
- Isolamento de falhas é apenas lógico: uma query destrutiva no schema `auth` pode afetar a disponibilidade global do banco. Em produção, `GRANT` mínimo e ferramentas de auditoria mitigam isso.
- Escala menos flexível: se o volume de dados do schema `videos` crescer muito mais rápido que os outros, toda a instância precisará ser escalada, não apenas o serviço responsável.
- Acoplamento implícito: todos os serviços dependem da mesma instância PostgreSQL. Uma falha na instância afeta todos os serviços simultaneamente.
- Migrar para bancos separados no futuro exigiria mover dados e remover as chaves estrangeiras cross-schema.

**Neutras:**
- Essa decisão reflete o escopo atual do projeto (hackathon acadêmico com cinco serviços). Em um ambiente de produção com equipes grandes e SLAs independentes por domínio, a migração para bancos separados pode ser justificada.
- Para produção na AWS, o uso de **Amazon RDS PostgreSQL** com réplicas de leitura endereça os problemas de disponibilidade e escala sem necessidade de particionar os dados por schema desde o início.
- O schema `auth` é o mais crítico (sem ele, nenhum serviço funciona). Em produção, replicação assíncrona garante fallback de leitura.
