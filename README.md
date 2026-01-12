# Video Processing Platform

A scalable microservices-based video processing platform that extracts frames from uploaded videos and provides them as downloadable ZIP files.

## Architecture

This project implements a Clean Architecture (Hexagonal) pattern with 5 microservices:

- **Auth Service** - User authentication with JWT tokens (Port 8081)
- **API Gateway** - Public-facing API for video operations (Port 8080)
- **Processing Worker** - FFmpeg video frame extraction (scalable workers)
- **Storage Service** - ZIP creation and download serving (Port 8082)
- **Notification Service** - Email notifications for processing completion/errors

### Technology Stack

**Backend:**
- Go 1.24+ with Chi Router + Uber FX
- PostgreSQL 15 (with schemas: auth, videos, notifications)
- Redis 7 (JWT cache, rate limiting)
- RabbitMQ 3 (async job queue)
- AWS S3 (video and frame storage)
- GORM (ORM)

**Frontend:**
- React 18+ with TypeScript
- Vite (build tool)
- TanStack Router (type-safe routing)
- TanStack Query (server state)
- shadcn/ui + TailwindCSS (UI components)
- Axios (HTTP client)
- Bun (package manager)

**Infrastructure:**
- Docker Compose
- Nginx (UI deployment)

## Project Structure

```
video-processing-platform/
├── services/
│   ├── auth/                  # Authentication service
│   ├── api-gateway/           # API Gateway service
│   ├── processing-worker/     # Video processing workers
│   ├── storage/               # Storage service
│   └── notification/          # Notification service
├── ui/                        # React frontend (Bun + Vite)
├── shared/pkg/                # Shared Go libraries
│   ├── database/              # PostgreSQL + Redis
│   ├── messaging/             # RabbitMQ
│   ├── auth/jwt/              # JWT utilities
│   ├── httpclient/            # HTTP client with retry
│   ├── storage/s3/            # S3 client
│   ├── rest/                  # REST utilities
│   ├── config/                # Configuration loader
│   └── logging/               # Logger
├── deployment/
│   └── docker-compose.yml     # Full stack deployment
├── scripts/
│   └── migrations/            # Database migrations
└── README.md
```

## Prerequisites

- Go 1.24+
- Bun (for UI)
- Docker & Docker Compose
- AWS Account (for S3)
- SMTP credentials (Gmail recommended)

## Quick Start

### 1. Clone the Repository

```bash
git clone <repository-url>
cd video-processing-platform
```

### 2. Configure Environment Variables

```bash
cp .env.example .env
```

Edit `.env` and update:
- `AWS_ACCESS_KEY_ID` - Your AWS access key
- `AWS_SECRET_ACCESS_KEY` - Your AWS secret key
- `S3_UPLOADS_BUCKET` - S3 bucket for uploaded videos
- `S3_PROCESSED_BUCKET` - S3 bucket for processed frames
- `JWT_SECRET` - Strong random secret for JWT
- `SMTP_USER` - Your email address
- `SMTP_PASSWORD` - Your email app password

### 3. Create S3 Buckets

```bash
# Using AWS CLI
aws s3 mb s3://video-platform-uploads
aws s3 mb s3://video-platform-processed
```

### 4. Start the Platform

```bash
cd deployment
docker-compose up -d
```

This will start:
- PostgreSQL (port 5432)
- Redis (port 6379)
- RabbitMQ (port 5672, management UI at 15672)
- Auth Service (port 8081)
- API Gateway (port 8080)
- Processing Workers (3 replicas)
- Storage Service (port 8082)
- Notification Service
- React UI (port 3000)
- Cleanup Cron Job

### 5. Access the Application

- **UI**: http://localhost:3000
- **API**: http://localhost:8080
- **RabbitMQ Management**: http://localhost:15672 (user: `video`, pass: `secret`)

## Development

### Backend Services

Each service follows Clean Architecture:

```
internal/{service}/
├── domain/           # Entities + Repository interfaces
├── usecase/          # Business logic
├── controller/       # Orchestration
├── presenter/        # Response formatting
└── infrastructure/   # API handlers, Persistence, External services
```

Build a service:
```bash
cd services/auth
go build -o bin/auth cmd/api/main.go
```

Run tests:
```bash
go test ./...
```

### Frontend (UI)

```bash
cd ui
bun install          # Install dependencies
bun run dev          # Start dev server
bun run build        # Build for production
bun run preview      # Preview production build
```

### Shared Libraries

Add dependencies to shared module:
```bash
cd shared
go get <package>
```

All services automatically access shared libraries through Go workspace.

## API Endpoints

### Auth Service (8081)

- `POST /register` - Register new user
- `POST /login` - Login (returns access + refresh tokens)
- `POST /refresh` - Refresh access token
- `POST /logout` - Logout (blacklist token)

### API Gateway (8080)

- `POST /videos/upload` - Upload video (auth required)
- `GET /videos` - List user's videos (auth required)
- `GET /videos/:id/status` - Get video status (auth required)
- `GET /videos/:id/download` - Download ZIP (auth required)

## Video Processing Flow

1. User uploads video → API Gateway
2. API Gateway → Streams to S3 → Creates DB record (status: PENDING)
3. Publishes job to RabbitMQ queue
4. Processing Worker → Consumes job → FFmpeg extraction @ 1fps
5. Worker → Uploads frames to S3 → Storage Service creates ZIP
6. Updates DB (status: COMPLETED)
7. Notification Service → Sends email to user
8. User downloads ZIP via presigned URL
9. After 15 days → Cron job deletes video + ZIP from S3 + DB

## Database Schema

### auth.users
- `id`, `username`, `email`, `password_hash`, `created_at`, `updated_at`

### auth.refresh_tokens
- `id`, `user_id`, `token`, `expires_at`, `created_at`

### videos.videos
- `id`, `user_id`, `filename`, `original_path`, `status`, `fps`, `frame_count`, `zip_path`, `error_message`, `created_at`, `started_at`, `completed_at`, `expires_at`

### notifications.notification_log
- `id`, `user_id`, `video_id`, `type`, `status`, `recipient`, `subject`, `error_message`, `sent_at`, `created_at`

## Monitoring

- **RabbitMQ**: http://localhost:15672
- **Logs**: `docker-compose logs -f <service-name>`
- **Database**: Connect to PostgreSQL on port 5432

## Scaling

### Processing Workers

Scale workers based on queue depth:
```bash
docker-compose up -d --scale processing-worker=10
```

### Database

For production, use managed PostgreSQL with read replicas.

### S3

S3 automatically scales. Consider CloudFront CDN for downloads.

## Security

- JWT tokens: 15 min access, 7 day refresh
- Passwords: bcrypt (cost 12)
- S3: Private buckets with presigned URLs (15 min expiry)
- Rate limiting: 10 uploads/hour per user
- File validation: Magic numbers, 500MB max, extension whitelist
- Data retention: Auto-delete after 15 days

## Troubleshooting

### Database connection failed
```bash
docker-compose logs postgres
docker-compose restart postgres
```

### RabbitMQ not working
```bash
docker-compose logs rabbitmq
# Check http://localhost:15672 for queue status
```

### S3 access denied
Verify AWS credentials and bucket policies in `.env`

### Video processing stuck
Check worker logs:
```bash
docker-compose logs processing-worker
```

## License

MIT

## Contributors

FIAP SOAT10 - Hackathon Project
