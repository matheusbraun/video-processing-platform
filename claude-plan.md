# Video Processing Platform - Implementation Plan

## Overview

Transform the monolithic projeto-fiapx video processing application into a scalable microservices architecture following Clean Architecture patterns from tc-fiap-customer and tc-fiap-order.

## System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              FRONTEND LAYER                                  │
│                                                                              │
│  ┌────────────────────────────────────────────────────────────────────┐     │
│  │  React UI (TanStack Router + shadcn/ui)                            │     │
│  │  - Login/Register (email/password)                                 │     │
│  │  - Upload Videos (drag-and-drop)                                   │     │
│  │  - List Videos (status: pending/processing/completed/failed)       │     │
│  │  - Download ZIPs                                                   │     │
│  └────────────────────────────────────────────────────────────────────┘     │
│                                  │                                           │
│                                  │ HTTP/REST (Port 8080)                     │
│                                  ▼                                           │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                          MICROSERVICES LAYER                                 │
│                                                                              │
│  ┌──────────────────┐          ┌────────────────────────────────┐           │
│  │  Auth Service    │◄─────────│     API Gateway Service        │           │
│  │   (Port 8081)    │  Validate│        (Port 8080)             │           │
│  │                  │   JWT    │                                │           │
│  │ - Register       │          │ - Upload Video (protected)     │           │
│  │ - Login          │          │ - List Videos (protected)      │           │
│  │ - Refresh Token  │          │ - Get Status (protected)       │           │
│  │ - Logout         │          │ - Download (protected)         │           │
│  └──────────────────┘          └────────────────────────────────┘           │
│          │                                    │         │                    │
│          │                                    │         │                    │
│          │                                    │         └─────────┐          │
│          ▼                                    ▼                   ▼          │
│  ┌──────────────────┐          ┌────────────────────┐  ┌─────────────────┐  │
│  │   PostgreSQL     │          │   RabbitMQ Queue   │  │ Storage Service │  │
│  │                  │          │                    │  │  (Port 8082)    │  │
│  │ Schemas:         │          │ Queue:             │  │                 │  │
│  │ - auth (users)   │          │ video.processing   │  │ - Create ZIP    │  │
│  │ - videos         │          │                    │  │ - Serve Download│  │
│  │ - notifications  │          └────────────────────┘  └─────────────────┘  │
│  └──────────────────┘                    │                       │           │
│                                           │                       │           │
│                                           ▼                       ▼           │
│                          ┌──────────────────────────┐    ┌──────────────┐   │
│                          │  Processing Worker Pool  │    │   AWS S3     │   │
│                          │  (10-50 replicas)        │    │              │   │
│                          │                          │    │ Buckets:     │   │
│                          │ - Consume jobs           │───▶│ - uploads    │   │
│                          │ - FFmpeg extraction      │    │ - processed  │   │
│                          │ - Upload frames to S3    │    │              │   │
│                          │ - Update DB status       │    │              │   │
│                          │ - Publish completion     │    │              │   │
│                          └──────────────────────────┘    └──────────────┘   │
│                                           │                                  │
│                                           │ Publish to notification.queue    │
│                                           ▼                                  │
│                          ┌──────────────────────────┐                        │
│                          │  Notification Service    │                        │
│                          │                          │                        │
│                          │ - Consume events         │                        │
│                          │ - Send email (SMTP/SES)  │                        │
│                          │ - Retry failed sends     │                        │
│                          └──────────────────────────┘                        │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                        INFRASTRUCTURE LAYER                                  │
│                                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │  PostgreSQL  │  │    Redis     │  │   RabbitMQ   │  │  Cron Job    │    │
│  │   (5432)     │  │    (6379)    │  │   (5672)     │  │              │    │
│  │              │  │              │  │   Mgmt: 15672│  │  Daily 2AM:  │    │
│  │ - Persistence│  │ - JWT cache  │  │              │  │  Delete old  │    │
│  │ - Migrations │  │ - Rate limit │  │ - Job queue  │  │  videos from │    │
│  │              │  │ - Sessions   │  │ - Pub/Sub    │  │  S3 + DB     │    │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘    │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘


VIDEO PROCESSING FLOW:
═══════════════════════

1. User uploads video (React UI → API Gateway)
                ↓
2. API Gateway: Validate JWT → Stream to S3 → Create DB record (status: PENDING)
                ↓
3. Publish message to RabbitMQ: { job_id, user_id, video_id, s3_path }
                ↓
4. Processing Worker: Consume → Download video → FFmpeg extract frames @ 1fps
                ↓
5. Upload frames to S3 → Storage Service: Create ZIP
                ↓
6. Update DB: status=COMPLETED, zip_path, frame_count
                ↓
7. Publish to notification queue → Notification Service: Send email
                ↓
8. User downloads ZIP (API Gateway → Storage Service → S3 presigned URL)
                ↓
9. After 15 days: Cron job deletes video + ZIP from S3 + DB


CLEAN ARCHITECTURE PATTERN (Each Service):
══════════════════════════════════════════

┌─────────────────────────────────────────┐
│         HTTP Request (Chi Router)       │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│      Infrastructure Layer               │
│  - API Controller (HTTP handlers)       │
│  - DTOs (Request/Response)              │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│      Presentation Layer                 │
│  - Controller (Orchestration)           │
│  - Presenter (Format output)            │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│      Application Layer                  │
│  - Use Cases (Business Logic)           │
│  - Commands (Input DTOs)                │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│      Domain Layer                       │
│  - Entities (Core objects)              │
│  - Repository Interfaces                │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│      Infrastructure Layer               │
│  - Repository Implementations (GORM)    │
│  - Database Clients                     │
│  - External Services (S3, SMTP)         │
└─────────────────────────────────────────┘
```

## Architecture Decision: Monorepo with 5 Microservices

### Repository Structure
```
video-processing-platform/
├── go.work                    # Go workspace (manages all modules)
├── services/
│   ├── auth/                  # Authentication service
│   ├── api-gateway/           # Video upload/list/download API
│   ├── processing-worker/     # FFmpeg video processing
│   ├── storage/               # File storage and ZIP creation
│   └── notification/          # Email notifications
├── ui/                        # React frontend application
│   ├── package.json
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── services/          # API client
│   │   └── App.tsx
│   ├── Dockerfile
│   └── nginx.conf
├── shared/pkg/                # Shared libraries
│   ├── database/              # PostgreSQL + Redis clients
│   ├── messaging/             # RabbitMQ abstractions
│   ├── auth/                  # JWT middleware
│   ├── httpclient/            # HTTP client with retry
│   ├── storage/               # S3 client
│   └── rest/                  # HTTP utilities
├── deployment/
│   ├── docker-compose.yml     # Local development & production
│   └── k8s/                   # Kubernetes manifests (future)
└── scripts/
    └── cleanup-old-videos.sh  # Cron job for 15-day retention
```

## Microservices Breakdown

### 1. Auth Service (Port 8081)
- **Responsibility**: User registration, login, JWT token management
- **Database**: PostgreSQL schema `auth` (users, refresh_tokens)
- **Endpoints**: POST /register, POST /login, POST /refresh, POST /logout
- **Dependencies**: PostgreSQL, Redis (token blacklist)

### 2. API Gateway (Port 8080)
- **Responsibility**: Public-facing API for video operations
- **Database**: PostgreSQL schema `videos` (videos table)
- **Endpoints**:
  - POST /videos/upload (auth required)
  - GET /videos (list user's videos, auth required)
  - GET /videos/:id/status (auth required)
  - GET /videos/:id/download (auth required)
- **Dependencies**: Auth Service, Storage Service, RabbitMQ, PostgreSQL, Redis

### 3. Processing Worker (No HTTP server)
- **Responsibility**: Consume jobs from queue, run FFmpeg extraction, upload frames
- **Database**: PostgreSQL schema `videos` (update status, frame_count)
- **Flow**: RabbitMQ consumer → FFmpeg extraction → Upload frames to S3 → Update DB → Publish completion event
- **Dependencies**: RabbitMQ, PostgreSQL, MinIO/S3
- **Scaling**: 10-50 replicas based on queue depth

### 4. Storage Service (Port 8082)
- **Responsibility**: Create ZIP from frames, serve downloads
- **Database**: PostgreSQL schema `videos` (update zip_path)
- **Endpoints**: Internal only (called by API Gateway)
- **Dependencies**: MinIO/S3, PostgreSQL

### 5. Notification Service (No HTTP server)
- **Responsibility**: Send email notifications on completion/error
- **Database**: PostgreSQL schema `notifications` (notification_log)
- **Flow**: RabbitMQ consumer → Send email via SMTP
- **Dependencies**: RabbitMQ, SMTP/SES, PostgreSQL

## Message Flow

```
User uploads video
    ↓
API Gateway (auth check)
    ↓ (save metadata, stream to S3)
    ↓
RabbitMQ queue: video.processing.queue
    ↓
Processing Worker (consume job)
    ↓ (FFmpeg extraction)
    ↓
Storage Service (create ZIP)
    ↓ (update DB: status=COMPLETED)
    ↓
RabbitMQ queue: video.notification.queue
    ↓
Notification Service (send email)
```

## Technology Stack

### Backend
- **Language**: Go 1.24+
- **Framework**: Chi Router v5 + Uber FX v1.23
- **Architecture**: Clean Architecture (Domain → UseCase → Controller → Infrastructure)
- **Database**: PostgreSQL 15 (single instance for Docker Compose)
- **Cache**: Redis 7
- **Queue**: RabbitMQ (single instance for Docker Compose)
- **Storage**: AWS S3
- **ORM**: GORM v1.30
- **Containerization**: Docker Compose

### Frontend
- **Framework**: React 18+ with TypeScript
- **Build Tool**: Vite
- **UI Library**: shadcn/ui (Radix UI + TailwindCSS)
- **HTTP Client**: Axios
- **Routing**: TanStack Router (type-safe routing)
- **State Management**: TanStack Query (for server state)
- **Deployment**: Nginx static server

## Database Schema

### Schema: auth
```sql
CREATE TABLE auth.users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE auth.refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES auth.users(id),
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Schema: videos
```sql
CREATE TABLE videos.videos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id INTEGER NOT NULL REFERENCES auth.users(id),
    filename VARCHAR(255) NOT NULL,
    original_path TEXT NOT NULL,
    status VARCHAR(20) NOT NULL, -- PENDING, PROCESSING, COMPLETED, FAILED
    fps INTEGER DEFAULT 1,
    frame_count INTEGER,
    zip_path TEXT,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    expires_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP + INTERVAL '15 days')
);

CREATE INDEX idx_user_status ON videos.videos(user_id, status);
CREATE INDEX idx_created_at ON videos.videos(created_at);
CREATE INDEX idx_expires_at ON videos.videos(expires_at) WHERE status = 'COMPLETED';
```

### Schema: notifications
```sql
CREATE TABLE notifications.notification_log (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES auth.users(id),
    video_id UUID REFERENCES videos.videos(id),
    type VARCHAR(20) NOT NULL, -- EMAIL, WEBHOOK
    status VARCHAR(20) NOT NULL, -- PENDING, SENT, FAILED
    recipient TEXT NOT NULL,
    subject TEXT,
    error_message TEXT,
    sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Shared Code Strategy

All services will use shared modules from `shared/pkg/`:

1. **database/postgres**: PostgreSQL connection with GORM (pattern from tc-fiap-order)
2. **database/redis**: Redis client for caching
3. **messaging/rabbitmq**: RabbitMQ publisher/consumer abstractions
4. **auth/jwt**: JWT middleware for route protection
5. **httpclient**: Resilient HTTP client with retry (from tc-fiap-order)
6. **rest**: Controller interface and response utilities (from tc-fiap-customer)
7. **storage/s3**: AWS S3 client with AWS SDK v2

## Clean Architecture Pattern (Per Service)

Following tc-fiap patterns:

```
internal/{service}/
├── domain/
│   ├── entities/           # Business objects (Video, User, Job)
│   └── repositories/       # Repository interfaces only
├── usecase/
│   ├── {action}/
│   │   ├── {action}_use_case.go      # Interface
│   │   └── {action}_use_case_impl.go # Implementation
│   └── commands/           # Input DTOs
├── controller/
│   ├── {service}_controller.go       # Interface
│   └── {service}_controller_impl.go  # Orchestrates use cases
├── presenter/
│   ├── {service}_presenter.go        # Interface
│   └── {service}_presenter_impl.go   # Format output
└── infrastructure/
    ├── api/
    │   ├── controller/     # HTTP handlers (Chi)
    │   └── dto/            # Request/response DTOs
    └── persistence/        # Repository implementations (GORM)
```

## Dependency Injection (Uber FX)

Each service will follow this pattern (from tc-fiap-order):

```go
// internal/app/app.go
func InitializeApp() *fx.App {
    return fx.New(
        fx.Provide(
            config.Load,
            postgres.NewPostgresDB,
            redis.NewRedisClient,
            
            // Repositories
            fx.Annotate(persistence.NewRepoImpl, fx.As(new(repositories.Repo))),
            
            // Use Cases
            fx.Annotate(usecase.NewUseCaseImpl, fx.As(new(usecase.UseCase))),
            
            // Controllers
            fx.Annotate(controller.NewControllerImpl, fx.As(new(controller.Controller))),
            
            chi.NewRouter,
        ),
        fx.Invoke(registerRoutes),
        fx.Invoke(startHTTPServer),
    )
}
```

## Scalability Strategy

### Processing Worker Auto-Scaling (HPA)
```yaml
minReplicas: 10
maxReplicas: 50
metrics:
  - type: External
    metric: rabbitmq_queue_messages_ready
    target: 10 messages per pod
  - type: Resource
    resource: cpu
    target: 80%
```

### API Gateway Auto-Scaling
```yaml
minReplicas: 5
maxReplicas: 20
metrics:
  - type: Resource
    resource: cpu
    target: 70%
```

## Critical Files from Reference Projects

1. **/Users/mbraun/dev/go/forks/tc-fiap-order/internal/app/app.go**
   - Uber FX dependency injection setup

2. **/Users/mbraun/dev/go/forks/tc-fiap-order/internal/order/usecase/addOrder/add_order_use_case_impl.go**
   - Use case implementation pattern

3. **/Users/mbraun/dev/go/forks/tc-fiap-order/internal/shared/httpclient/http_client.go**
   - HTTP client for inter-service communication

4. **/Users/mbraun/dev/go/forks/tc-fiap-customer/internal/customer/domain/repositories/customer_respository.go**
   - Repository interface pattern

5. **/Users/mbraun/dev/go/forks/projeto-fiapx/main.go**
   - Current logic to decompose

## Migration Mapping

| Current Code (main.go) | New Service | Component |
|------------------------|-------------|-----------|
| `handleVideoUpload()` | API Gateway | UploadVideoUseCase |
| `isValidVideoFile()` | API Gateway | ValidateVideoUseCase |
| `processVideo()` | Processing Worker | ExtractFramesUseCase |
| FFmpeg execution | Processing Worker | FFmpegService (infrastructure) |
| `createZipFile()` | Storage Service | CreateZipUseCase |
| `handleDownload()` | Storage Service | DownloadVideoUseCase |
| `handleStatus()` | API Gateway | ListUserVideosUseCase |
| User auth (none) | Auth Service | LoginUseCase, RegisterUseCase |
| Error handling | Notification Service | SendNotificationUseCase |

## Implementation Steps

### Week 1-2: Infrastructure Setup
1. Create monorepo structure with go.work
2. Setup shared/pkg modules (database, messaging, auth, S3, etc)
3. Create docker-compose.yml with PostgreSQL, Redis, RabbitMQ
4. Configure AWS S3 buckets (video-platform-uploads, video-platform-processed)
5. Create database schemas and migration scripts
6. Setup React UI boilerplate:
   - Vite + TypeScript
   - TanStack Router with type-safe routes
   - shadcn/ui components (Button, Card, Input, etc)
   - TailwindCSS configuration
7. Setup CI/CD pipeline (GitHub Actions)

### Week 3: Auth Service
1. Implement domain entities (User)
2. Implement repositories (UserRepository)
3. Implement use cases:
   - RegisterUseCase (email/password, bcrypt hashing)
   - LoginUseCase (returns JWT access + refresh tokens)
   - RefreshTokenUseCase
   - LogoutUseCase (blacklist token in Redis)
4. Implement JWT middleware
5. Create HTTP API controllers with CORS for React UI
6. Unit + integration tests

### Week 4-5: API Gateway
1. Implement domain entities (Video)
2. Implement repositories (VideoRepository)
3. Implement use cases:
   - UploadVideoUseCase (save to S3, create DB record, queue job)
   - ListUserVideosUseCase
   - GetVideoStatusUseCase
4. Integrate with Auth Service (JWT validation)
5. Create HTTP API controllers
6. Unit + integration tests

### Week 6-7: Processing Worker
1. Implement RabbitMQ consumer
2. Implement ExtractFramesUseCase (FFmpeg wrapper)
3. Implement frame upload to S3
4. Update video status in DB
5. Publish completion event to notification queue
6. Error handling with Dead Letter Queue
7. Load testing with multiple workers

### Week 8: Storage Service
1. Implement CreateZipUseCase
2. Implement DownloadVideoUseCase (presigned URLs)
3. S3/MinIO integration
4. Unit + integration tests

### Week 9: Notification Service
1. Implement RabbitMQ consumer for notification events
2. Implement SendEmailUseCase (SMTP/SES)
3. Email templates (HTML + plain text)
4. Retry logic with exponential backoff
5. Unit + integration tests

### Week 9-10: React UI
1. Setup project structure (routes, components, services)
2. Configure TanStack Router:
   - Route definitions with type safety
   - Layout routes (auth layout, protected layout)
   - Route guards for authentication
3. Implement authentication routes:
   - /login - Login form with shadcn/ui components
   - /register - Registration form
   - JWT token management (localStorage + refresh)
4. Implement video routes:
   - /upload - Upload page with drag-and-drop (shadcn/ui dropzone)
   - /videos - Video list with shadcn/ui Cards and Badges
   - /videos/$videoId - Download page
5. shadcn/ui components to use:
   - Button, Input, Card, Badge, Alert, Dropdown, Dialog, Toast
6. API client with Axios + TanStack Query
7. Error handling with shadcn/ui Toast notifications

### Week 10: Video Cleanup Job
1. Create cron container with cleanup script
2. Script logic:
   - Query videos WHERE expires_at < NOW()
   - Delete from S3 (original + ZIP)
   - Delete from database
3. Add to docker-compose.yml (runs daily at 2 AM)

### Week 11: Testing & Hardening
1. Backend E2E tests (Godog BDD)
2. Frontend E2E tests (Playwright with TanStack Router)
3. Load testing (k6) - simulate concurrent uploads
4. Security audit:
   - File validation (magic numbers, size limits)
   - JWT expiration and refresh flow
   - CORS configuration
   - S3 bucket policies (private access only)
5. Performance tuning
6. Monitoring setup (Prometheus + Grafana dashboards)

### Week 12: Production Deployment
1. Test full stack with docker-compose
2. User acceptance testing
3. Documentation:
   - README with setup instructions
   - Architecture diagram
   - API documentation (Swagger)
4. Video presentation (10 min):
   - Architecture walkthrough
   - Live demo (upload → process → download → cleanup)
   - Monitoring dashboards

## Monitoring & Observability

### Key Metrics
- Queue depth (RabbitMQ)
- Processing time per video (P50, P95, P99)
- Error rate per service
- API latency
- Database connection pool usage

### Prometheus + Grafana Dashboards
1. System Overview (request rate, error rate, latency)
2. Video Processing (queue depth, processing time, success/fail rate)
3. Database Health (connections, query duration)
4. Infrastructure (CPU/memory/disk per service)

## Security

- **Authentication**: Email/password with JWT tokens (15 min access + 7 day refresh)
- **Registration**: Open registration (anyone can create account)
- **Authorization**: User-scoped queries (WHERE user_id = :user_id)
- **Passwords**: bcrypt (cost 12)
- **Rate Limiting**: 10 uploads/hour per user (Redis)
- **File Validation**: Magic number check, size limit (500MB max), extension whitelist
- **S3 Security**: 
  - Private buckets (no public access)
  - IAM role for service access
  - Presigned URLs for downloads (15 min expiry)
- **CORS**: API Gateway allows requests from UI origin only
- **Data Retention**: Automatic deletion after 15 days (compliance)

## Docker Compose Configuration

### Services included:
- PostgreSQL 15 (single instance)
- Redis 7 (single instance)
- RabbitMQ 3 with management UI (single instance)
- All 5 Go microservices
- React UI (Nginx)

### Storage:
- AWS S3 (requires AWS credentials in .env)
- Buckets: `video-platform-uploads`, `video-platform-processed`

### Cleanup Job:
- Cron container running daily cleanup script
- Deletes videos older than 15 days from S3 and database

## React UI Structure

```
ui/
├── package.json
├── tsconfig.json
├── vite.config.ts
├── tailwind.config.js
├── components.json          # shadcn/ui configuration
├── src/
│   ├── main.tsx
│   ├── routes/
│   │   ├── __root.tsx       # Root layout
│   │   ├── _auth/           # Auth layout (guest only)
│   │   │   ├── login.tsx
│   │   │   └── register.tsx
│   │   ├── _protected/      # Protected layout (auth required)
│   │   │   ├── index.tsx    # Dashboard/videos list
│   │   │   ├── upload.tsx
│   │   │   └── videos.$videoId.tsx
│   │   └── index.tsx        # Landing page
│   ├── components/
│   │   ├── ui/              # shadcn/ui components
│   │   │   ├── button.tsx
│   │   │   ├── card.tsx
│   │   │   ├── input.tsx
│   │   │   ├── badge.tsx
│   │   │   ├── toast.tsx
│   │   │   └── ...
│   │   ├── Header.tsx
│   │   ├── VideoCard.tsx
│   │   └── UploadZone.tsx
│   ├── lib/
│   │   ├── api.ts           # Axios client with interceptors
│   │   ├── auth.ts          # Auth utilities
│   │   ├── utils.ts         # cn() and other utils
│   │   └── queryClient.ts   # TanStack Query setup
│   ├── hooks/
│   │   ├── useAuth.ts       # Auth hook with TanStack Query
│   │   └── useVideos.ts     # Videos hook with TanStack Query
│   ├── types/
│   │   ├── auth.types.ts
│   │   └── video.types.ts
│   └── utils/
│       └── formatters.ts    # Date, file size formatters
├── public/
│   └── index.html
├── nginx.conf               # Production nginx config
└── Dockerfile               # Multi-stage: build + nginx
```

## Docker Compose Port Mapping

| Service | Internal Port | External Port | Purpose |
|---------|--------------|---------------|---------|
| UI (Nginx) | 80 | 3000 | React frontend |
| API Gateway | 8080 | 8080 | Public API |
| Auth Service | 8080 | 8081 | Auth endpoints |
| Storage Service | 8080 | 8082 | Internal only (via API Gateway) |
| PostgreSQL | 5432 | 5432 | Database |
| Redis | 6379 | 6379 | Cache |
| RabbitMQ | 5672 | 5672 | Message queue |
| RabbitMQ Management | 15672 | 15672 | Queue monitoring UI |

## Environment Variables (.env file)

```bash
# Database
DATABASE_URL=postgres://videoadmin:secret@postgres:5432/video_platform

# Redis
REDIS_URL=redis://redis:6379

# RabbitMQ
RABBITMQ_URL=amqp://video:secret@rabbitmq:5672/

# AWS S3
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
S3_UPLOADS_BUCKET=video-platform-uploads
S3_PROCESSED_BUCKET=video-platform-processed

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# Service URLs
AUTH_SERVICE_URL=http://auth-service:8080
API_GATEWAY_URL=http://api-gateway:8080
STORAGE_SERVICE_URL=http://storage-service:8080

# SMTP (for notifications)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# Frontend
VITE_API_URL=http://localhost:8080
```

## Next Steps After Plan Approval

1. Create repository: `video-processing-platform/`
2. Initialize Go workspace with `go work init`
3. Create all service directories with boilerplate
4. Setup React UI with Vite
5. Create docker-compose.yml with all services
6. Create database migration scripts
7. Setup AWS S3 buckets
8. Begin Auth Service implementation
