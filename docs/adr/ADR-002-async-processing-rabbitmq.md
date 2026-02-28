# ADR-002: RabbitMQ para processamento assíncrono de vídeos

**Status:** Aceito
**Data:** 2026-02-01

---

## Contexto

A extração de frames de um vídeo é realizada pelo FFmpeg e pode levar de alguns segundos a vários minutos, dependendo da duração e da resolução do vídeo. Três abordagens foram consideradas:

1. **Processamento síncrono no API Gateway**: o endpoint `/videos/upload` aguardaria a conclusão do FFmpeg antes de responder ao cliente.
2. **Chamada HTTP direta do API Gateway para um serviço de worker**: o gateway publicaria a requisição de forma assíncrona via HTTP, mas o worker seria acessado por chamada direta.
3. **Fila de mensagens (RabbitMQ)**: o gateway publica um evento na fila e workers independentes o consomem.

A opção 1 foi descartada imediatamente: timeouts de HTTP são tipicamente de dezenas de segundos, e o cliente ficaria bloqueado. A opção 2 melhora a desvinculação do tempo de resposta, mas mantém o acoplamento direto entre gateway e worker — se o worker estiver sobrecarregado ou indisponível, o gateway falharia.

---

## Decisão

Adotar **RabbitMQ** como broker de mensagens AMQP para desacoplar o API Gateway dos Processing Workers.

**Implementação:**
- O API Gateway publica uma mensagem JSON na fila `video_processing_queue` após salvar o vídeo no S3 e criar o registro no banco.
- Três réplicas do Processing Worker consomem a fila de forma competitiva (cada mensagem é processada por exatamente um worker).
- `QoS prefetch=1`: cada worker aceita apenas uma mensagem por vez, garantindo que a fila distribua carga igualmente e que um worker lento não trave mensagens não processadas.
- Em caso de falha no processamento, o worker publica um evento de notificação de erro para a `notification_queue` e registra a falha no banco.
- O Notification Service consome `notification_queue` de forma independente para envio de e-mails.

**Formato da mensagem em `video_processing_queue`:**
```json
{
  "video_id": "uuid",
  "user_id": 42,
  "original_path": "uploads/uuid/original.mp4"
}
```

---

## Consequências

**Positivas:**
- O endpoint de upload responde em milissegundos (202 Accepted), independentemente da duração do processamento.
- Workers escalam horizontalmente de forma independente: `docker-compose up --scale processing-worker=N`.
- Falhas no worker não afetam o gateway; mensagens ficam na fila e são reprocessadas quando o worker se recupera.
- Back-pressure natural: se a fila encher além do suportado, o gateway continua publicando, mas os workers processam no seu ritmo.
- O mesmo padrão é reutilizado para notificações, tornando a arquitetura consistente.

**Negativas:**
- O status do vídeo é **eventualmente consistente**: o cliente precisa fazer polling de `GET /videos/{id}/status` para saber quando o processamento terminou. A UI implementa auto-polling para resolver isso.
- Adiciona RabbitMQ como dependência de infraestrutura, com seus próprios requisitos de operação e monitoramento.
- Mensagens duplicadas são possíveis (RabbitMQ é at-least-once). O sistema trata isso via idempotência no banco (UPDATE com condição de status).

**Neutras:**
- O RabbitMQ Management UI (`:15672`) permite inspecionar filas e monitorar profundidade em tempo real.
- O Alertmanager dispara alerta quando a fila ultrapassa 1.000 mensagens, sinalizando necessidade de mais workers.
