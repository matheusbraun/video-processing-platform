# Architecture Decision Records (ADRs)

Este diretório contém os registros de decisões arquiteturais (ADRs) da Plataforma de Processamento de Vídeo.

Cada ADR documenta **uma decisão de design significativa**: o contexto que a motivou, o que foi decidido e quais são as consequências esperadas. O formato adotado é o [Nygard ADR](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions).

---

## Índice

| # | Título | Status | Data |
|---|--------|--------|------|
| [ADR-001](ADR-001-clean-architecture.md) | Adoção da Clean Architecture em todos os serviços Go | Aceito | 2026-02-01 |
| [ADR-002](ADR-002-async-processing-rabbitmq.md) | RabbitMQ para processamento assíncrono de vídeos | Aceito | 2026-02-01 |
| [ADR-003](ADR-003-jwt-redis-session.md) | JWT com blacklist via Redis para gerenciamento de sessão | Aceito | 2026-02-01 |
| [ADR-004](ADR-004-s3-video-storage.md) | AWS S3 para armazenamento de vídeos e frames | Aceito | 2026-02-01 |
| [ADR-005](ADR-005-go-workspace-monorepo.md) | Go workspace (`go.work`) para monorepo com bibliotecas compartilhadas | Aceito | 2026-02-01 |
| [ADR-006](ADR-006-uber-fx-dependency-injection.md) | Uber FX para injeção de dependências nos serviços Go | Aceito | 2026-02-01 |
| [ADR-007](ADR-007-multi-schema-postgresql.md) | PostgreSQL com múltiplos schemas em vez de bancos separados por serviço | Aceito | 2026-02-01 |

---

## Como ler um ADR

Cada arquivo segue a estrutura:

- **Título** — Decisão em uma frase
- **Status** — `Proposto`, `Aceito`, `Depreciado` ou `Substituído por ADR-XXX`
- **Data** — Quando a decisão foi tomada
- **Contexto** — O problema ou situação que motivou a decisão
- **Decisão** — O que foi decidido e como foi implementado
- **Consequências** — Impactos positivos, negativos e neutros da decisão

## Como adicionar um novo ADR

1. Copie o template de qualquer ADR existente
2. Numere sequencialmente (`ADR-008-...`)
3. Defina o status como `Proposto`
4. Adicione uma linha na tabela acima
5. Ao ser aceito pela equipe, mude o status para `Aceito`
