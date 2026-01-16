package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for a service
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	// Queue metrics (for workers)
	QueueMessagesProcessed  *prometheus.CounterVec
	QueueProcessingDuration *prometheus.HistogramVec
	QueueMessagesInFlight   prometheus.Gauge

	// Database metrics
	DBQueriesTotal    *prometheus.CounterVec
	DBQueryDuration   *prometheus.HistogramVec
	DBConnectionsOpen prometheus.Gauge

	// Business metrics
	VideosTotal             *prometheus.CounterVec
	VideoProcessingDuration *prometheus.HistogramVec
	FramesExtracted         prometheus.Counter
	EmailsSent              *prometheus.CounterVec
}

// NewMetrics creates a new Metrics instance for a service
func NewMetrics(serviceName string) *Metrics {
	return &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"method", "endpoint", "status_code"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"method", "endpoint"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
		),

		// Queue metrics
		QueueMessagesProcessed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "queue_messages_processed_total",
				Help: "Total number of queue messages processed",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"queue", "status"},
		),
		QueueProcessingDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "queue_processing_duration_seconds",
				Help:    "Queue message processing duration in seconds",
				Buckets: []float64{.1, .5, 1, 5, 10, 30, 60, 120, 300},
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"queue"},
		),
		QueueMessagesInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "queue_messages_in_flight",
				Help: "Current number of queue messages being processed",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
		),

		// Database metrics
		DBQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_queries_total",
				Help: "Total number of database queries",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"operation", "status"},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: []float64{.001, .005, .01, .05, .1, .5, 1},
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"operation"},
		),
		DBConnectionsOpen: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_open",
				Help: "Current number of open database connections",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
		),

		// Business metrics
		VideosTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "videos_total",
				Help: "Total number of videos by status",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"status"},
		),
		VideoProcessingDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "video_processing_duration_seconds",
				Help:    "Video processing duration in seconds",
				Buckets: []float64{1, 5, 10, 30, 60, 120, 300, 600, 1200},
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"operation"},
		),
		FramesExtracted: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "frames_extracted_total",
				Help: "Total number of frames extracted from videos",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
		),
		EmailsSent: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "emails_sent_total",
				Help: "Total number of emails sent",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			[]string{"status"},
		),
	}
}
