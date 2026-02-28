# API Gateway Service

The API Gateway is the main public-facing service that handles video upload operations and orchestrates processing workflows.

## Features

- **Video Upload**: Multi-part file upload with streaming to S3
- **Video Listing**: Paginated list of user videos
- **Status Tracking**: Real-time processing status
- **Download**: Presigned S3 URLs for ZIP downloads
- **JWT Authentication**: Secure endpoints with JWT validation
- **API Documentation**: Interactive Swagger UI

## Running Locally

### Prerequisites
- Go 1.24+
- PostgreSQL (or use Docker Compose)
- Redis (or use Docker Compose)
- RabbitMQ (or use Docker Compose)
- AWS S3 credentials

### Environment Variables

```bash
SERVER_PORT=8080
DATABASE_URL=postgres://user:password@localhost:5432/video_platform?sslmode=disable
REDIS_URL=redis://localhost:6379
RABBITMQ_URL=amqp://user:pass@localhost:5672/
AUTH_SERVICE_URL=http://localhost:8081
STORAGE_SERVICE_URL=http://localhost:8082
JWT_SECRET=your-secret-key
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
S3_UPLOADS_BUCKET=video-platform-uploads
S3_PROCESSED_BUCKET=video-platform-processed
```

### Start the Service

#### Using Docker Compose (Recommended)

```bash
# Start all services
make up

# View logs
make logs

# Stop services
make down
```

#### Run Locally (Requires infrastructure)

```bash
# Start only infrastructure (DB, Redis, RabbitMQ)
make infra

# Generate Swagger docs and run
make swagger
go run cmd/server/main.go
```

## Common Commands

```bash
# Start all services
make up

# View API Gateway logs
make logs

# Restart API Gateway
make restart

# Rebuild after code changes
make rebuild

# Scale processing workers
make scale WORKERS=5

# Stop all services
make down

# Show all commands
make help
```

## API Documentation

### Swagger UI

The service provides interactive API documentation using Swagger/OpenAPI.

**Access:** http://localhost:8080/swagger/index.html

### Key Features:
- 📖 **Interactive Documentation**: See all available endpoints with detailed descriptions
- 🧪 **Try It Out**: Test endpoints directly from the browser
- 🔐 **Authentication Support**: Click "Authorize" button and enter JWT token
- 📊 **Schema Definitions**: View request/response models and examples
- 🎯 **Parameter Details**: See all required and optional parameters

### How to Use Swagger UI

1. **Access the UI**: Navigate to http://localhost:8080/swagger/index.html

2. **Authenticate**:
   - Click the "Authorize" button (🔓 icon)
   - Enter your JWT token in the format: `Bearer <your-token>`
   - Click "Authorize" and then "Close"

3. **Test an Endpoint**:
   - Expand the endpoint you want to test
   - Click "Try it out"
   - Fill in the required parameters
   - Click "Execute"
   - See the response below

### Available Endpoints

#### POST /api/v1/videos/upload
Upload a video file for processing (max 500MB).

**Request:**
- `video` (form-data file): The video file

**Response:**
```json
{
  "video_id": "uuid",
  "filename": "example.mp4",
  "status": "PENDING"
}
```

#### GET /api/v1/videos
List all videos for the authenticated user.

**Query Parameters:**
- `limit` (optional, default: 20): Number of videos to return
- `offset` (optional, default: 0): Number of videos to skip

**Response:**
```json
{
  "videos": [...],
  "total": 100,
  "limit": 20,
  "offset": 0,
  "has_more": true
}
```

#### GET /api/v1/videos/{id}/status
Get processing status for a specific video.

**Path Parameters:**
- `id`: Video ID (UUID)

**Response:**
```json
{
  "video_id": "uuid",
  "filename": "example.mp4",
  "status": "COMPLETED",
  "frame_count": 120,
  "created_at": "2026-02-06T10:00:00Z",
  "completed_at": "2026-02-06T10:05:00Z"
}
```

#### GET /api/v1/videos/{id}/download
Get a presigned URL to download the processed frames ZIP.

**Path Parameters:**
- `id`: Video ID (UUID)

**Response:**
```json
{
  "download_url": "https://s3.amazonaws.com/...",
  "filename": "frames.zip",
  "expires_in": 3600
}
```

## Regenerating Swagger Docs

When you modify handlers or add new endpoints:

```bash
make swagger
```

## Development

```bash
# Generate Swagger docs
make swagger

# Build and run
make build
make run

# Run tests
make test

# Clean artifacts
make clean
```

## Architecture

The service follows Clean Architecture with these layers:

- **domain**: Entities and repository interfaces
- **usecase**: Business logic (upload, list, status, download)
- **controller**: Orchestrates use cases
- **presenter**: Formats responses
- **infrastructure**:
  - **api/controller**: HTTP handlers
  - **api/dto**: Request/response DTOs
  - **persistence**: Database repositories

## Health Check

```bash
curl http://localhost:8080/health
# Returns: OK
```

## Testing

```bash
# Run all tests
go test ./...

## Testing

```bash
make test
```

## Troubleshooting

### Swagger UI Not Loading

Regenerate docs: `make swagger`

### Authentication Errors

Ensure JWT token format: `Bearer <token>`
