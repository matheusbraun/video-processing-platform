.PHONY: swagger run build test clean help
.PHONY: up down logs restart infra rebuild scale

COMPOSE_DIR = ../../deployment

# ============================================================================
# Development
# ============================================================================

swagger:
	swag init -g cmd/server/main.go --output docs --parseDependency --parseInternal

run:
	go run cmd/server/main.go

build:
	go build -o bin/api-gateway cmd/server/main.go

test:
	go test ./... -v

clean:
	rm -rf bin/ docs/ coverage.out coverage.html

# ============================================================================
# Docker Compose
# ============================================================================

up:
	cd $(COMPOSE_DIR) && docker-compose up -d
	@echo "\n✓ Services started!"
	@echo "  UI: http://localhost:3000"
	@echo "  API: http://localhost:8080"
	@echo "  Swagger: http://localhost:8080/swagger/index.html\n"

down:
	cd $(COMPOSE_DIR) && docker-compose down

logs:
	cd $(COMPOSE_DIR) && docker-compose logs -f api-gateway

restart:
	cd $(COMPOSE_DIR) && docker-compose restart api-gateway

infra:
	cd $(COMPOSE_DIR) && docker-compose up -d postgres redis rabbitmq

rebuild: swagger
	cd $(COMPOSE_DIR) && docker-compose up -d --build api-gateway

scale:
	cd $(COMPOSE_DIR) && docker-compose up -d --scale processing-worker=$(WORKERS)

# ============================================================================
# Help
# ============================================================================

help:
	@echo "Development:"
	@echo "  make swagger    Generate Swagger docs"
	@echo "  make run        Run locally"
	@echo "  make build      Build binary"
	@echo "  make test       Run tests"
	@echo ""
	@echo "Docker Compose:"
	@echo "  make up         Start all services"
	@echo "  make down       Stop all services"
	@echo "  make logs       View API Gateway logs"
	@echo "  make restart    Restart API Gateway"
	@echo "  make infra      Start only DB/Redis/RabbitMQ"
	@echo "  make rebuild    Rebuild API Gateway"
	@echo "  make scale WORKERS=5  Scale workers"
	@echo ""

