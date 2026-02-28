# ADR-001: Adoção da Clean Architecture em todos os serviços Go

**Status:** Aceito
**Data:** 2026-02-01

---

## Contexto

A plataforma é composta por cinco serviços Go com responsabilidades distintas (autenticação, gateway, processamento, armazenamento e notificação). Desde o início, havia o requisito de alta testabilidade — a meta era atingir acima de 80% de cobertura por serviço — e de flexibilidade para trocar detalhes de infraestrutura (banco de dados, fila, provedores de armazenamento) sem impactar as regras de negócio.

Dois problemas são recorrentes em projetos Go sem uma estrutura clara:
1. **Acoplamento de infraestrutura ao negócio**: lógica de negócio acaba diretamente acoplada ao GORM, ao RabbitMQ ou ao AWS SDK, tornando-a impossível de testar sem subir todos esses serviços.
2. **Ausência de fronteiras claras**: qualquer parte do código pode chamar qualquer outra, tornando o rastreamento de responsabilidades difícil e as mudanças arriscadas.

---

## Decisão

Todos os serviços Go adotam a **Clean Architecture** (também conhecida como Arquitetura Hexagonal ou Ports & Adapters), com a seguinte estrutura de diretórios:

```
internal/{service}/
├── domain/           # Entidades e interfaces de repositório (sem dependências externas)
├── usecase/          # Regras de negócio (dependem apenas do domínio)
├── controller/       # Traduz entrada HTTP/AMQP → comando de caso de uso
├── presenter/        # Formata a saída do caso de uso → resposta HTTP/JSON
└── infrastructure/   # Implementações concretas (GORM, S3, RabbitMQ, SMTP)
```

**Regra de dependência**: as camadas internas (`domain`, `usecase`) nunca importam camadas externas (`infrastructure`). A infraestrutura implementa as interfaces definidas no domínio via injeção de dependência (ver ADR-006).

**Consequência prática:** os casos de uso são testados com mocks das interfaces de repositório (gerados via Mockery), sem necessidade de banco de dados ou fila real. As integrações concretas são testadas separadamente com testcontainers.

---

## Consequências

**Positivas:**
- Cobertura de testes média de **87,7%** entre os serviços; API Gateway atingiu **100%** nos casos de uso.
- As regras de negócio são independentes de frameworks: é possível trocar o Chi Router, o GORM ou o AWS SDK sem alterar a lógica de negócio.
- Novos colaboradores têm um modelo mental claro de onde cada tipo de código deve ficar.
- Casos de uso são unidades de teste naturais — cada caso de uso tem uma suíte de testes dedicada.

**Negativas:**
- Mais arquivos e interfaces do que uma abordagem simples de handler → repository direto.
- Exige disciplina da equipe para não "vazar" dependências de infraestrutura para camadas internas.
- O padrão pode ser excessivo para serviços muito simples (ex.: Cleanup Cron), mas foi mantido por consistência.

**Neutras:**
- A estrutura de diretórios é previsível entre todos os serviços, facilitando a navegação no monorepo.
