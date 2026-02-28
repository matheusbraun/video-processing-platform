# ADR-005: Go workspace (`go.work`) para monorepo com bibliotecas compartilhadas

**Status:** Aceito
**Data:** 2026-02-01

---

## Contexto

A plataforma tem cinco serviços Go com responsabilidades distintas, mas com necessidades de infraestrutura em comum: conexão com PostgreSQL, cliente Redis, cliente RabbitMQ, cliente S3, geração e validação de JWT, logging estruturado e métricas Prometheus.

Sem uma estratégia de compartilhamento de código, cada serviço duplicaria essas implementações — violando o princípio DRY e tornando mudanças (como atualizar a versão do AWS SDK) um trabalho de cinco pull requests em vez de um.

Três abordagens foram consideradas:

1. **Copiar código entre serviços**: rápido no início, mas divergências surgem rapidamente. Bugs corrigidos em um serviço permanecem em outros.
2. **Publicar um módulo Go privado no registro**: requer configuração de autenticação, versionamento semântico e publicação a cada mudança. Overhead elevado para uma plataforma mantida por uma equipe pequena.
3. **Go workspace (`go.work`)**: recurso nativo do Go 1.18+ que permite referenciar módulos locais diretamente, sem publicação. O compilador resolve as importações do workspace automaticamente.

---

## Decisão

Adotar um **monorepo com Go workspace** (`go.work` na raiz do repositório):

```
go.work                          # Declara o workspace com todos os módulos
go.work.sum                      # Checksums unificados
shared/                          # Módulo compartilhado: github.com/org/video-platform/shared
└── pkg/
    ├── config/
    ├── logging/
    ├── database/postgres/
    ├── database/redis/
    ├── messaging/rabbitmq/
    ├── auth/jwt/
    ├── storage/s3/
    ├── httpclient/
    ├── metrics/
    └── rest/
services/
├── auth/                        # Módulo independente: github.com/org/video-platform/auth
├── api-gateway/                 # Módulo independente
├── processing-worker/           # Módulo independente
├── storage/                     # Módulo independente
└── notification/                # Módulo independente
```

Cada serviço importa `shared/pkg/...` como se fosse um módulo externo. O `go.work` diz ao compilador para resolver esse import a partir do diretório local `./shared`, sem precisar de uma versão publicada.

**Restrição importante no CI/CD:** todos os comandos `go build`, `go test` e `docker build` precisam ser executados a partir da raiz do repositório (onde o `go.work` está) para que o workspace seja reconhecido. Os Dockerfiles copiam o monorepo completo antes de compilar o serviço específico.

---

## Consequências

**Positivas:**
- Uma única mudança no pacote `shared/pkg/logging/` propaga para todos os serviços sem necessidade de versionamento ou publicação.
- Atualizações de dependências (ex.: AWS SDK, GORM) são feitas uma vez no `go.work.sum` e compartilhadas.
- O desenvolvimento local é simples: `go test ./...` na raiz executa todos os testes de todos os módulos.
- Cada serviço continua sendo um módulo Go independente com seu próprio `go.mod` — é possível extraí-lo do monorepo no futuro se necessário.

**Negativas:**
- Os Dockerfiles precisam copiar o repositório inteiro (ou ao menos o `go.work` + `shared/` + o serviço específico) antes de compilar. Isso aumenta o contexto de build do Docker em comparação com um serviço totalmente isolado.
- A ausência de versionamento explícito do módulo compartilhado significa que mudanças incompatíveis podem quebrar outros serviços imediatamente — exige mais cuidado na revisão de código de `shared/`.
- Ferramentas que não reconhecem Go workspaces (algumas versões de IDEs ou linters) podem ter dificuldade de análise.

**Neutras:**
- Go workspaces são o mecanismo oficial do Go para este cenário desde a versão 1.18.
- A estrutura facilita a criação de novos serviços: basta criar um novo diretório em `services/`, adicionar ao `go.work` e importar `shared/pkg/` normalmente.
