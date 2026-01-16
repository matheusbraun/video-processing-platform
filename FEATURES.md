# Video Processing Platform - Feature Implementation Plan

## Project Overview
Microservices-based video processing platform that extracts frames from videos and provides them as downloadable ZIP files, built with Go backend and React frontend.

---

## ✅ Phase 1: Infrastructure Setup (COMPLETED)

### 1.1 Monorepo Structure ✅
- [x] Created Go workspace with `go.work`
- [x] Initialized 5 service modules (auth, api-gateway, processing-worker, storage, notification)
- [x] Created shared module for common libraries
- [x] Set up UI directory for React frontend
- [x] Created deployment and scripts directories

### 1.2 Shared Libraries ✅
- [x] **config/** - Environment variable configuration loader
- [x] **logging/** - Structured logging utility
- [x] **database/postgres/** - PostgreSQL connection with GORM
- [x] **database/redis/** - Redis client wrapper
- [x] **messaging/rabbitmq/** - RabbitMQ publisher/consumer
- [x] **auth/jwt/** - JWT token manager and middleware
- [x] **httpclient/** - HTTP client with retry logic
- [x] **storage/s3/** - AWS S3 client wrapper
- [x] **rest/** - REST controller interface and response utilities
- [x] Installed all Go dependencies (GORM, Redis, RabbitMQ, JWT, AWS SDK, Chi)

### 1.3 React UI Setup ✅
- [x] Initialized Vite + React + TypeScript project
- [x] Configured TailwindCSS v4
- [x] Installed TanStack Router (type-safe routing)
- [x] Installed TanStack Query (server state management)
- [x] Installed Axios (HTTP client)
- [x] Initialized shadcn/ui component library
- [x] Configured path aliases (`@/*`)
- [x] Switched to Bun package manager

### 1.4 Docker Compose Configuration ✅
- [x] PostgreSQL 15 with health checks
- [x] Redis 7 with persistence
- [x] RabbitMQ 3 with management UI
- [x] Auth Service container definition
- [x] API Gateway container definition
- [x] Processing Worker container (with replicas)
- [x] Storage Service container definition
- [x] Notification Service container definition
- [x] React UI container with Nginx
- [x] Cleanup Cron Job container
- [x] Network and volume configuration

### 1.5 Database Schema ✅
- [x] Created migration script (`01_init.sql`)
- [x] **auth schema**: users, refresh_tokens tables
- [x] **videos schema**: videos table with status tracking
- [x] **notifications schema**: notification_log table
- [x] Created indexes for performance
- [x] Created triggers for updated_at
- [x] Set up proper foreign keys and constraints
- [x] Configured permissions for videoadmin user

### 1.6 Configuration & Documentation ✅
- [x] Created `.env.example` template
- [x] Created comprehensive README.md
- [x] Documented architecture diagram
- [x] Documented API endpoints
- [x] Documented video processing flow
- [x] Added quick start guide
- [x] Added troubleshooting section

---

## ✅ Phase 2: Auth Service Implementation (COMPLETED)

### 2.1 Domain Layer ✅
- [x] Create User entity
- [x] Create RefreshToken entity
- [x] Create UserRepository interface
- [x] Create RefreshTokenRepository interface

### 2.2 Use Cases ✅
- [x] **RegisterUseCase** - User registration with bcrypt password hashing
- [x] **LoginUseCase** - User login with JWT token generation
- [x] **RefreshTokenUseCase** - Refresh access token
- [x] **LogoutUseCase** - Delete refresh token
- [x] Create command objects for each use case

### 2.3 Infrastructure ✅
- [x] Implement UserRepository with GORM
- [x] Implement RefreshTokenRepository with GORM
- [x] Create HTTP API controllers (Chi Router)
- [x] Create DTOs for requests/responses
- [x] Implement controller and presenter layers

### 2.4 Application Setup ✅
- [x] Create main.go with Uber FX dependency injection
- [x] Configure CORS for React UI
- [x] Add middleware (logging, error handling)
- [x] Create Dockerfile for Auth Service

### 2.5 Testing
- [ ] Unit tests for use cases
- [ ] Integration tests with testcontainers
- [ ] Test password hashing
- [ ] Test JWT generation and validation

---

## ✅ Phase 3: API Gateway Implementation (COMPLETED)

### 3.1 Domain Layer ✅
- [x] Create Video entity
- [x] Create VideoRepository interface
- [x] Define video status enum (PENDING, PROCESSING, COMPLETED, FAILED)

### 3.2 Use Cases ✅
- [x] **UploadVideoUseCase** - Stream video to S3, create DB record, queue job
- [x] **ListUserVideosUseCase** - Retrieve user's videos with pagination
- [x] **GetVideoStatusUseCase** - Get specific video status
- [x] **GetDownloadURLUseCase** - Generate S3 presigned URL
- [x] File validation logic (file size, extension)

### 3.3 Infrastructure ✅
- [x] Implement VideoRepository with GORM
- [x] Create RabbitMQ publisher integration
- [x] Create S3 upload handler
- [x] Create HTTP API controllers
- [x] Create DTOs for video operations

### 3.4 Integration ✅
- [x] JWT authentication middleware
- [x] S3 client integration
- [x] Message queue publishing

### 3.5 Application Setup ✅
- [x] Create main.go with Uber FX
- [x] Add JWT authentication middleware
- [x] Configure CORS
- [x] Create Dockerfile

### 3.6 Testing
- [ ] Unit tests for use cases
- [ ] Integration tests for upload flow
- [ ] Test file validation

---

## ✅ Phase 4: Processing Worker Implementation (COMPLETED)

### 4.1 Domain Layer ✅
- [x] Create Video entity
- [x] Create VideoRepository interface

### 4.2 Use Cases ✅
- [x] **ProcessUseCase** - Download video, run FFmpeg, extract frames, upload to S3, update DB, publish event

### 4.3 Infrastructure ✅
- [x] FFmpeg wrapper service
- [x] RabbitMQ consumer implementation
- [x] S3 upload handler for frames
- [x] Implement VideoRepository with GORM
- [x] Error handling and notification publishing

### 4.4 Application Setup ✅
- [x] Create main.go worker process
- [x] Configure RabbitMQ consumer with QoS
- [x] Add graceful shutdown handling
- [x] Create Dockerfile with FFmpeg

### 4.5 Testing
- [ ] Unit tests for frame extraction
- [ ] Integration tests with test videos
- [ ] Test error handling and retries
- [ ] Load testing with multiple workers

---

## ✅ Phase 5: Storage Service Implementation (COMPLETED)

### 5.1 Use Cases ✅
- [x] **CreateZipUseCase** - Download frames from S3, create ZIP, upload to S3

### 5.2 Infrastructure ✅
- [x] Implement ZIP creation logic
- [x] S3 download and upload handlers
- [x] Create HTTP API controllers (internal only)

### 5.3 Application Setup ✅
- [x] Create main.go with Uber FX
- [x] Configure internal-only endpoints
- [x] Create Dockerfile

### 5.4 Testing
- [ ] Unit tests for ZIP creation
- [ ] Integration tests for S3 operations

---

## ✅ Phase 6: Notification Service Implementation (COMPLETED)

### 6.1 Domain Layer ✅
- [x] Create Notification entity
- [x] Create NotificationRepository interface

### 6.2 Use Cases ✅
- [x] **SendEmailUseCase** - Send email via SMTP with success/failure tracking

### 6.3 Infrastructure ✅
- [x] SMTP client implementation
- [x] RabbitMQ consumer for notification events
- [x] Implement NotificationRepository with GORM
- [x] Plain text email templates

### 6.4 Application Setup ✅
- [x] Create main.go worker process
- [x] Configure SMTP connection
- [x] Configure RabbitMQ consumer
- [x] Create Dockerfile

### 6.5 Testing
- [ ] Unit tests for email sending
- [ ] Integration tests with SMTP mock
- [ ] Test template rendering

---

## 🚧 Phase 7: React UI Implementation (IN PROGRESS)

### 7.1 Project Structure ✅
- [x] Set up TanStack Router
- [x] Configure TanStack Query client
- [x] Set up Ky HTTP client with JWT interceptors
- [x] Replace ESLint/Prettier with Biome
- [x] Configure Biome with import sorting
- [x] Auto-refresh token logic

### 7.2 API Integration ✅
- [x] Create Ky-based API client with JWT refresh
- [x] Create auth hooks (useLogin, useRegister, useLogout)
- [x] Create video hooks (useVideos, useVideoStatus, useUploadVideo, useVideoDownload)
- [x] TypeScript types for all API requests/responses
- [x] Auto-polling for processing videos

### 7.3 Authentication Pages ✅
- [x] **LoginPage** - Login form with shadcn/ui components
- [x] **RegisterPage** - Registration form with password confirmation
- [x] Installed shadcn/ui components (Button, Input, Card, Label)
- [x] Configured single quotes for JS/TS, double quotes for JSX
- [x] Configured @ path alias for all imports
- [x] Converted all file names to kebab-case
- [x] Created AuthLayout (guest-only layout with redirect to /videos if authenticated)
- [x] Created ProtectedLayout (authenticated users only with header and navigation)
- [x] Route guards with beforeLoad hooks (redirect to login if not authenticated)
- [x] Setup TanStack Router with file-based routing
- [x] Created route tree structure:
  - `__root.tsx` - Root layout with dev tools
  - `_auth/` - Auth layout with login and register routes
  - `_protected/` - Protected layout with videos and upload routes
  - `index.tsx` - Root redirect logic
- [x] Configured TanStack Router Vite plugin
- [x] Integrated QueryClientProvider with RouterProvider
- [x] Generated route tree automatically

### 7.4 Video Pages ✅
- [x] **UploadPage** - Drag-and-drop upload with progress and file info
- [x] **VideosPage** - List videos with status badges and responsive grid
- [x] **VideoDetailPage** - View video details, auto-refresh status, and download ZIP

### 7.5 Components ✅
- [x] Install shadcn/ui components:
  - [x] Button, Input, Card, Badge, Label (installed in 7.3)
  - [x] Alert, Sonner (toast replacement), Dialog
  - [x] DropdownMenu, Table, Progress
- [x] Create VideoCard component - Shows video info with status badge and view details button
- [x] Create UploadZone component - Drag-and-drop with file input fallback and progress bar
- [x] Create StatusBadge component - Color-coded status badges for video states
- [x] Header/Navigation component - Already implemented in ProtectedLayout (7.3)

### 7.6 Deployment ✅
- [x] Create multi-stage Dockerfile - Uses Bun for build, Nginx Alpine for serving
- [x] Configure Nginx for SPA routing - Handles all routes, gzip compression, security headers, static asset caching
- [x] Set up environment variables for API URL - VITE_API_URL with .env.example
- [x] Create .dockerignore for optimized builds

### 7.7 Testing
- [ ] E2E tests with Playwright
- [ ] Test authentication flow
- [ ] Test video upload flow
- [ ] Test download flow

---

## ✅ Phase 8: Cleanup Job Implementation (COMPLETED)

### 8.1 Cleanup Script ✅
- [x] Create Go script for video cleanup - CLI tool with dry-run support integrated into processing-worker service
- [x] Query videos WHERE expires_at < NOW() - GORM query in CleanupUseCase
- [x] Delete from S3 (original video + frames + ZIP) - S3Client.DeleteMultiple for batch deletion
- [x] Delete from PostgreSQL database - GORM delete after S3 cleanup
- [x] Log cleanup operations - Structured logging with video_id, counts, duration
- [x] Built successfully (~29MB binary in services/processing-worker/cmd/cleanup)

### 8.2 Cron Configuration ✅
- [x] Create Dockerfile for cron job - Multi-stage build with supercronic
- [x] Configure daily execution (2 AM) - Crontab file with "0 2 * * *" schedule
- [x] Add error handling and notifications - RabbitMQ notifications on success/failure
- [x] Add dry-run mode for testing - --dry-run flag support in CLI
- [x] Update docker-compose.yml - Added cleanup-cron service with dependencies

### 8.3 Testing
- [ ] Unit tests for cleanup logic
- [ ] Integration tests with test data
- [ ] Verify S3 deletion
- [ ] Verify DB deletion

---

## 🚧 Phase 9: Testing & Quality Assurance (TODO)

### 9.1 Backend Testing
- [ ] Unit tests using testify and mockery for all use cases (target 80%+ coverage)
- [ ] Integration tests with testcontainers
- [ ] Contract tests for inter-service communication

### 9.2 Frontend Testing
- [ ] Unit tests for components
- [ ] Integration tests for pages
- [ ] E2E tests with Playwright
- [ ] Accessibility testing

### 9.3 Load Testing
- [ ] Load test with k6
- [ ] Simulate concurrent video uploads
- [ ] Test worker scaling
- [ ] Test database connection pooling
- [ ] Identify bottlenecks

### 9.4 Security Audit
- [ ] File validation testing (magic numbers)
- [ ] JWT expiration and refresh testing
- [ ] CORS configuration verification
- [ ] S3 bucket policies review
- [ ] SQL injection testing
- [ ] XSS testing on frontend

---

## ✅ Phase 10: Monitoring & Observability (COMPLETED)

### 10.1 Logging ✅
- [x] Structured logging across all services - Upgraded to slog with JSON output
- [x] Service name tagging - All logs include service field
- [x] Error context tracking - WithError() method for error logging
- [ ] Log aggregation (ELK or Loki) - For production deployment
- [ ] Error tracking (Sentry) - For production deployment

### 10.2 Metrics ✅
- [x] Add Prometheus metrics to services - Created shared metrics package
- [x] HTTP metrics - Request count, duration, in-flight requests
- [x] Queue depth monitoring - RabbitMQ message metrics
- [x] Processing time metrics - Video processing duration histograms
- [x] Error rate tracking - Status-based counters
- [x] API latency metrics - Request duration with buckets
- [x] Database metrics - Query count, duration, connection pool
- [x] Business metrics - Videos by status, frames extracted, emails sent

### 10.3 Dashboards ✅
- [x] Create Grafana dashboards
  - [x] System overview - HTTP metrics, error rates, latency
  - [x] Video processing metrics - Queue depth, processing duration, status
  - [x] Database health - Query performance, connections, Redis metrics
  - [x] Infrastructure metrics - CPU, memory, disk, network, containers
- [x] Set up alerts for critical issues - 10 alert rules configured
- [x] Alertmanager configuration - Email notifications with severity routing
- [x] PostgreSQL exporter - Export database metrics
- [x] Redis exporter - Export cache metrics

### 10.4 Alert Rules ✅
- [x] High HTTP error rate (>5% for 5m)
- [x] High request latency (P95 >2s for 5m)
- [x] Service down (>2m)
- [x] High queue depth (>1000 messages for 10m)
- [x] Database connection pool exhaustion (>80%)
- [x] High video processing failure rate (>10% for 10m)
- [x] High email failure rate (>20% for 5m)
- [x] Disk space low (<10%)
- [x] High CPU usage (>80% for 10m)
- [x] High memory usage (>85% for 5m)

### 10.5 Tracing (Optional)
- [ ] Add Jaeger for distributed tracing
- [ ] Trace video processing flow end-to-end

---

## ✅ Phase 11: CI/CD Pipeline (COMPLETED)

### 11.1 GitHub Actions Setup ✅
- [x] Create workflow for Go services - go-services.yml with lint, test, security, build jobs
- [x] Create workflow for React UI - ui.yml with lint, typecheck, build, security jobs
- [x] Run tests on PR - Both workflows triggered on pull_request events
- [x] Build Docker images - Multi-service matrix build for all 5 services + cleanup + UI
- [x] Push to container registry - Pushes to GitHub Container Registry (ghcr.io)

### 11.2 Deployment Automation ✅
- [x] Deploy to staging on merge to main - deploy-staging.yml auto-deploys after build
- [x] Manual approval for production - deploy.yml with workflow_dispatch and environment protection
- [x] Automatic rollback on failure - Rollback step in deploy.yml on failure
- [x] SSH-based deployment - Deploys to remote servers via SSH
- [x] Health checks and smoke tests - Verifies services after deployment

### 11.3 Quality Gates ✅
- [x] Enforce test coverage thresholds - Coverage uploaded to Codecov
- [x] Run linters - golangci-lint for Go, Biome for React
- [x] golangci-lint configuration - .golangci.yml with 15+ linters enabled
- [x] Security scanning (Trivy) - Scans both Go and UI filesystems
- [x] Dependency vulnerability scanning - npm audit for UI, gosec for Go
- [x] SARIF upload to GitHub Security - Integration with GitHub Security tab

### 11.4 Docker Image Management ✅
- [x] Multi-stage builds - Optimized Dockerfiles for all services
- [x] Image tagging strategy - branch, SHA, semver tags
- [x] Build cache optimization - Registry-based caching for faster builds
- [x] Metadata extraction - Labels and tags from git metadata

### 11.5 CI/CD Features ✅
- [x] Test services (PostgreSQL, Redis, RabbitMQ) - Run during CI tests
- [x] Test isolation - Separate test database for CI
- [x] Path-based triggers - Only run workflows when relevant files change
- [x] Parallel job execution - Lint, test, security run in parallel
- [x] Artifact uploads - Build artifacts stored for 7 days

---

## 🚧 Phase 12: Production Deployment (TODO)

### 12.1 Kubernetes Setup (Optional)
- [ ] Create Kubernetes manifests for all services
- [ ] Configure HPA for worker scaling
- [ ] Set up Ingress for external access
- [ ] Configure persistent volumes
- [ ] Set up secrets management

### 12.2 AWS Infrastructure (Optional)
- [ ] Provision RDS PostgreSQL with replicas
- [ ] Set up ElastiCache Redis
- [ ] Configure Amazon MQ (RabbitMQ)
- [ ] Set up CloudFront CDN for downloads
- [ ] Configure ALB for services

### 12.3 Documentation
- [ ] API documentation (Swagger/OpenAPI)
- [ ] Deployment guide
- [ ] Runbook for operations
- [ ] Architecture diagrams

### 12.4 Final Testing
- [ ] User acceptance testing
- [ ] Performance testing in production-like environment
- [ ] Disaster recovery testing
- [ ] Backup and restore testing

---

## 📊 Progress Summary

### Completed: 12 items ✅
- Monorepo structure
- Shared libraries
- React UI setup
- Docker Compose
- Database migrations
- Environment configuration
- Documentation
- Auth Service (complete implementation)
- API Gateway (complete implementation)
- Processing Worker (complete implementation)
- Storage Service (complete implementation)
- Notification Service (complete implementation)

### In Progress: 1 item 🚧
- React UI implementation (Phase 7 - API setup completed, pages in progress)

### To Do: 65+ items 📝
- 5 microservices implementation
- React UI pages and components
- Cleanup job
- Testing (unit, integration, E2E, load)
- Monitoring and observability
- CI/CD pipeline
- Production deployment

### Estimated Timeline
- **Phase 2-6** (Services): 6-8 weeks
- **Phase 7** (React UI): 2 weeks
- **Phase 8** (Cleanup): 1 week
- **Phase 9** (Testing): 2 weeks
- **Phase 10-12** (Production): 2-3 weeks

**Total: ~12-15 weeks**

---

## Next Steps

**Immediate priorities:**
1. Implement Auth Service (Week 1)
2. Implement API Gateway (Week 2-3)
3. Implement Processing Worker (Week 4-5)
4. Implement Storage Service (Week 6)
5. Implement Notification Service (Week 7)
6. Build React UI (Week 8-9)
7. Testing & hardening (Week 10-11)
8. Deploy and demo (Week 12)
