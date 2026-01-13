FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY shared/ shared/
COPY services/api-gateway/ services/api-gateway/

WORKDIR /app/services/api-gateway

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /api-gateway cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /api-gateway .

EXPOSE 8080

CMD ["./api-gateway"]
