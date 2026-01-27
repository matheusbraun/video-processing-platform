FROM golang:1.24-alpine AS builder

RUN apk add --no-cache ffmpeg

WORKDIR /app

COPY go.work go.work.sum ./
COPY shared/ shared/
COPY services/auth/ services/auth/
COPY services/api-gateway/ services/api-gateway/
COPY services/storage/ services/storage/
COPY services/notification/ services/notification/
COPY services/processing-worker/ services/processing-worker/

WORKDIR /app/services/processing-worker

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /processing-worker cmd/worker/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates ffmpeg

WORKDIR /root/

COPY --from=builder /processing-worker .

CMD ["./processing-worker"]
