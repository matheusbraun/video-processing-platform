FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY shared/ shared/
COPY services/auth/ services/auth/
COPY services/api-gateway/ services/api-gateway/
COPY services/storage/ services/storage/
COPY services/notification/ services/notification/
COPY services/processing-worker/ services/processing-worker/

WORKDIR /app/services/auth

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /auth-service .

EXPOSE 8080

CMD ["./auth-service"]
