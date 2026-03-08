FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY shared/ shared/
COPY services/auth/ services/auth/
COPY services/api-gateway/ services/api-gateway/
COPY services/storage/ services/storage/
COPY services/notification/ services/notification/
COPY services/processing-worker/ services/processing-worker/

WORKDIR /app/services/api-gateway

RUN go mod download

# Generate Swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@latest \
    && swag init -g cmd/server/main.go --output docs --parseDependency --parseInternal

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /api-gateway cmd/server/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -u 10001 appuser

WORKDIR /app

# Copy binary
COPY --from=builder /api-gateway ./api-gateway

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

CMD ["./api-gateway"]
