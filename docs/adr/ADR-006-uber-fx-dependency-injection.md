# ADR-006: Uber FX para injeção de dependências nos serviços Go

**Status:** Aceito
**Data:** 2026-02-01

---

## Contexto

Cada serviço Go da plataforma possui um grafo de dependências não trivial. O API Gateway, por exemplo, precisa de:

- Conexão com PostgreSQL (GORM)
- Cliente Redis
- Cliente RabbitMQ (publisher)
- Cliente S3
- JWTManager
- VideoRepository (depende de PostgreSQL)
- UploadVideoUseCase (depende de VideoRepository, S3Client, RabbitMQPublisher)
- ListUserVideosUseCase (depende de VideoRepository)
- GetVideoStatusUseCase (depende de VideoRepository)
- GetDownloadURLUseCase (depende de VideoRepository, S3Client)
- HTTPServer Chi (depende de todos os use cases + JWTManager)

Sem um framework de injeção de dependências, o `main.go` precisa instanciar cada objeto na ordem correta, passando dependências manualmente. Isso cria um `main.go` frágil e difícil de manter — qualquer nova dependência exige alterar a cadeia de construção manualmente.

Duas abordagens foram avaliadas:

1. **Wiring manual no `main.go`**: simples de entender, sem dependências externas. Torna-se um "telescoping constructor" conforme o serviço cresce. Sem gerenciamento de lifecycle (Start/Stop ordenado).
2. **Uber FX**: framework de injeção de dependências baseado em reflexão desenvolvido pelo Uber, com gerenciamento de lifecycle. Valida o grafo de dependências na inicialização e falha de forma explicativa se algo estiver faltando.

---

## Decisão

Adotar **Uber FX** como framework de injeção de dependências em todos os serviços Go.

**Padrão de uso:**
- Cada componente (repositório, use case, controller, servidor HTTP) expõe um construtor (`NewXxx`) que recebe suas dependências como parâmetros
- No `main.go`, `fx.New()` recebe `fx.Provide(NewXxx)` para cada componente
- `fx.Invoke()` inicia os serviços que precisam ser inicializados mas não são dependência de ninguém (ex.: o servidor HTTP)
- Hooks `fx.Hook{OnStart, OnStop}` garantem graceful startup e shutdown

**Exemplo simplificado:**
```go
fx.New(
    fx.Provide(
        config.Load,
        postgres.NewConnection,
        redis.NewClient,
        rabbitmq.NewPublisher,
        s3.NewClient,
        jwt.NewManager,
        repository.NewVideoRepository,
        usecase.NewUploadVideoUseCase,
        usecase.NewListUserVideosUseCase,
        server.NewHTTPServer,
    ),
    fx.Invoke(func(s *server.HTTPServer, lc fx.Lifecycle) {
        lc.Append(fx.Hook{
            OnStart: func(ctx context.Context) error { return s.Start() },
            OnStop:  func(ctx context.Context) error { return s.Stop(ctx) },
        })
    }),
)
```

Se qualquer dependência estiver faltando ou tiver tipos incompatíveis, o FX falha na inicialização com uma mensagem clara indicando qual dependência não foi satisfeita.

---

## Consequências

**Positivas:**
- O grafo de dependências é **validado em tempo de inicialização**: erros de wiring são detectados antes do serviço aceitar requisições.
- Gerenciamento de lifecycle automático: o FX chama os hooks `OnStart` e `OnStop` na ordem correta (respeitando o grafo), garantindo graceful shutdown.
- Adicionar uma nova dependência a um use case não exige alterar o `main.go` — basta atualizar o construtor do use case.
- O código de cada componente é independente do FX (construtores normais do Go), facilitando os testes unitários.

**Negativas:**
- Usa reflexão em Go, o que adiciona tempo de startup (tipicamente dezenas de milissegundos — aceitável para serviços de longa duração).
- A curva de aprendizado para desenvolvedores não familiarizados com frameworks de DI pode ser inicialmente alta.
- Erros de tipo no grafo de dependências só são detectados em tempo de execução (inicialização), não em tempo de compilação — embora a mensagem de erro do FX seja descritiva.

**Neutras:**
- O padrão é consistente em todos os cinco serviços, tornando a estrutura do `main.go` previsível e fácil de replicar ao criar novos serviços.
- O FX é amplamente utilizado em produção no Uber e em empresas de grande porte, com manutenção ativa.
