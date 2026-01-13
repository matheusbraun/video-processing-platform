FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY shared/ shared/
COPY services/storage/ services/storage/

WORKDIR /app/services/storage

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /storage-service cmd/server/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /storage-service .

EXPOSE 8080

CMD ["./storage-service"]
