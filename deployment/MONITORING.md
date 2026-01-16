# Video Platform - Monitoring & Observability

This document describes the monitoring and observability setup for the Video Processing Platform.

## Overview

The platform uses a modern observability stack:
- **Prometheus** - Metrics collection and storage
- **Grafana** - Metrics visualization and dashboards
- **Alertmanager** - Alert routing and notifications
- **Exporters** - PostgreSQL and Redis metrics exporters
- **slog** - Structured JSON logging in all services

## Components

### Prometheus (Port 9090)
- Scrapes metrics from all services every 15 seconds
- Stores time-series data with 15-day retention (default)
- Evaluates alerting rules
- Accessible at: http://localhost:9090

**Monitored Services:**
- auth-service:8080/metrics
- api-gateway:8080/metrics
- processing-worker:8080/metrics
- storage-service:8080/metrics
- notification-service:8080/metrics
- postgres-exporter:9187
- redis-exporter:9121
- rabbitmq:15692/metrics

### Grafana (Port 3001)
- Visualizes metrics from Prometheus
- Pre-configured dashboards for system overview, video processing, database, and infrastructure
- Accessible at: http://localhost:3001
- Default credentials: admin/admin (change via GRAFANA_USER/GRAFANA_PASSWORD env vars)

**Pre-configured Dashboards:**
1. **System Overview** - HTTP request rates, error rates, latency, service status
2. **Video Processing** - Queue depth, processing duration, videos by status, frames extracted
3. **Database Health** - Query performance, connection pool, PostgreSQL and Redis metrics
4. **Infrastructure** - CPU, memory, disk, network usage for all containers

### Alertmanager (Port 9093)
- Receives alerts from Prometheus
- Routes alerts based on severity (critical/warning)
- Sends email notifications
- Accessible at: http://localhost:9093

**Alert Routing:**
- **Critical alerts** - Immediate email with [CRITICAL] prefix
- **Warning alerts** - Email with [WARNING] prefix
- Grouped by alertname and service
- 12-hour repeat interval

### PostgreSQL Exporter (Port 9187)
- Exports PostgreSQL database metrics
- Monitors connections, queries, database size, etc.

### Redis Exporter (Port 9121)
- Exports Redis cache metrics
- Monitors memory usage, commands, keys, etc.

## Metrics Collected

### HTTP Metrics
- `http_requests_total` - Total HTTP requests by method, endpoint, status code
- `http_request_duration_seconds` - Request duration histogram
- `http_requests_in_flight` - Current number of requests being processed

### Queue Metrics
- `queue_messages_processed_total` - Total messages processed by queue and status
- `queue_processing_duration_seconds` - Message processing duration histogram
- `queue_messages_in_flight` - Current number of messages being processed

### Database Metrics
- `db_queries_total` - Total database queries by operation and status
- `db_query_duration_seconds` - Query duration histogram
- `db_connections_open` - Current number of open database connections

### Business Metrics
- `videos_total` - Total videos by status (PENDING, PROCESSING, COMPLETED, FAILED)
- `video_processing_duration_seconds` - Video processing duration histogram
- `frames_extracted_total` - Total frames extracted from videos
- `emails_sent_total` - Total emails sent by status

## Alert Rules

### Critical Alerts
- **ServiceDown** - Service unavailable for >2 minutes
- **DiskSpaceLow** - Less than 10% disk space remaining

### Warning Alerts
- **HighHTTPErrorRate** - >5% error rate for 5 minutes
- **HighRequestLatency** - P95 latency >2s for 5 minutes
- **HighQueueDepth** - >1000 messages pending for 10 minutes
- **DatabaseConnectionPoolNearlyExhausted** - >80% pool usage for 5 minutes
- **HighVideoProcessingFailureRate** - >10% failure rate for 10 minutes
- **HighEmailFailureRate** - >20% failure rate for 5 minutes
- **HighCPUUsage** - >80% CPU usage for 10 minutes
- **HighMemoryUsage** - >85% memory usage for 5 minutes

## Configuration

### Environment Variables

Add to `.env` file:

```bash
# Grafana
GRAFANA_USER=admin
GRAFANA_PASSWORD=your-secure-password

# Alertmanager (for email notifications)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
ALERT_EMAIL=admin@example.com
```

### Starting the Monitoring Stack

```bash
cd deployment
docker-compose up -d prometheus grafana alertmanager postgres-exporter redis-exporter
```

### Accessing Dashboards

1. Open Grafana: http://localhost:3001
2. Login with credentials (default: admin/admin)
3. Navigate to Dashboards → Browse
4. Select one of the pre-configured dashboards

### Viewing Alerts

1. **Prometheus Alerts**: http://localhost:9090/alerts
2. **Alertmanager**: http://localhost:9093
3. **Email notifications** sent to ALERT_EMAIL

## Adding Metrics to Services

To add Prometheus metrics to a service:

1. **Import the metrics package:**
```go
import "github.com/video-platform/shared/pkg/metrics"
```

2. **Initialize metrics in main.go:**
```go
m := metrics.NewMetrics("service-name")
```

3. **Add HTTP middleware:**
```go
router.Use(m.HTTPMiddleware)
```

4. **Expose metrics endpoint:**
```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

router.Handle("/metrics", promhttp.Handler())
```

5. **Record custom metrics:**
```go
// Increment counter
m.VideosTotal.WithLabelValues("COMPLETED").Inc()

// Observe histogram
m.VideoProcessingDuration.WithLabelValues("extract_frames").Observe(duration.Seconds())

// Set gauge
m.QueueMessagesInFlight.Inc()
defer m.QueueMessagesInFlight.Dec()
```

## Structured Logging

All services use `slog` for structured JSON logging:

```go
import "github.com/video-platform/shared/pkg/logging"

logger := logging.NewLogger("service-name")

// Basic logging
logger.Info("Processing video", "video_id", videoID, "user_id", userID)
logger.Error("Failed to process", "error", err, "video_id", videoID)

// With additional context
logger.With("request_id", requestID).Info("Request completed")

// With error context
logger.WithError(err).Error("Database connection failed")
```

**Log Format:**
```json
{
  "time": "2024-01-16T10:30:00Z",
  "level": "INFO",
  "msg": "Processing video",
  "service": "processing-worker",
  "video_id": "123e4567-e89b-12d3-a456-426614174000",
  "user_id": 42
}
```

## Troubleshooting

### Prometheus Not Scraping Metrics
1. Check service is exposing `/metrics` endpoint
2. Verify service is running: `docker ps`
3. Check Prometheus targets: http://localhost:9090/targets
4. Review Prometheus logs: `docker logs video-platform-prometheus`

### Grafana Dashboard Not Showing Data
1. Verify Prometheus datasource: Grafana → Configuration → Data Sources
2. Check Prometheus is collecting metrics: http://localhost:9090/graph
3. Verify time range in dashboard (default: last 1 hour)

### Alerts Not Firing
1. Check alert rules: http://localhost:9090/alerts
2. Verify Alertmanager configuration: `docker logs video-platform-alertmanager`
3. Check email SMTP settings in alertmanager.yml
4. Review alert thresholds in prometheus-alerts.yml

### No Email Notifications
1. Verify SMTP credentials in `.env`
2. Check Alertmanager logs: `docker logs video-platform-alertmanager`
3. Test SMTP connection manually
4. For Gmail, use App Password (not regular password)

## Best Practices

1. **Set appropriate alert thresholds** - Adjust based on your SLAs
2. **Monitor dashboard regularly** - Check system overview daily
3. **Review alert history** - Tune thresholds to reduce false positives
4. **Use structured logging** - Always include context (user_id, video_id, etc.)
5. **Track business metrics** - Videos processed, frames extracted, etc.
6. **Set up log aggregation** - For production, use ELK or Loki
7. **Enable distributed tracing** - Optional Jaeger integration for complex debugging

## Metrics Retention

- **Prometheus**: 15 days (default)
- **Grafana**: Dashboards stored in grafana_data volume
- **Logs**: Not aggregated (Docker container logs only)

For production, consider:
- Increasing Prometheus retention (--storage.tsdb.retention.time flag)
- Setting up remote storage (Thanos, Cortex)
- Implementing log aggregation (ELK, Loki)

## Further Reading

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Alertmanager Documentation](https://prometheus.io/docs/alerting/latest/alertmanager/)
- [Go slog Package](https://pkg.go.dev/log/slog)
