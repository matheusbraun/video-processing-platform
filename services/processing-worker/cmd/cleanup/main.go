package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mbraun/video-processing-platform/shared/pkg/messaging/rabbitmq"
	"github.com/video-platform/services/processing-worker/internal/usecase/cleanup"
	"github.com/video-platform/shared/pkg/config"
	"github.com/video-platform/shared/pkg/database/postgres"
	"github.com/video-platform/shared/pkg/logging"
	"github.com/video-platform/shared/pkg/storage/s3"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "Run in dry-run mode (no actual deletions)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := logging.NewLogger("cleanup")

	db, err := postgres.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	s3Client, err := s3.NewS3Client(
		cfg.AWSRegion,
		cfg.AWSAccessKeyID,
		cfg.AWSSecretAccessKey,
		cfg.S3UploadsBucket,
	)
	if err != nil {
		log.Fatalf("Failed to create S3 client: %v", err)
	}

	cleanupUseCase := cleanup.NewCleanupUseCaseImpl(db, s3Client, *logger, *dryRun)

	ctx := context.Background()
	startTime := time.Now()

	logger.Info("Starting video cleanup job", "dry_run", *dryRun)

	result, err := cleanupUseCase.CleanupExpiredVideos(ctx)
	duration := time.Since(startTime)

	if err != nil {
		logger.Error("Cleanup job failed", "error", err, "duration_seconds", duration.Seconds())
		sendNotification(cfg, logger, "FAILED", err, duration, result)
		os.Exit(1)
	}

	logger.Info("Cleanup job completed",
		"videos_deleted", result.VideosDeleted,
		"s3_objects_deleted", result.S3ObjectsDeleted,
		"duration_seconds", duration.Seconds(),
		"dry_run", *dryRun,
	)

	// Send success notification if any videos were deleted
	if result.VideosDeleted > 0 {
		sendNotification(cfg, logger, "SUCCESS", nil, duration, result)
	}
}

func sendNotification(cfg *config.Config, logger *logging.Logger, status string, err error, duration time.Duration, result *cleanup.CleanupResult) {
	publisher, pubErr := rabbitmq.NewPublisher(cfg.RabbitMQURL)
	if pubErr != nil {
		logger.Error("Failed to create RabbitMQ publisher for notification", "error", pubErr)
		return
	}
	defer publisher.Close()

	var subject, body string
	if status == "SUCCESS" {
		subject = "Video Cleanup Job Completed Successfully"
		body = fmt.Sprintf(
			"Cleanup job completed:\n\n"+
				"Videos Deleted: %d\n"+
				"S3 Objects Deleted: %d\n"+
				"Duration: %.2f seconds\n"+
				"Timestamp: %s",
			result.VideosDeleted,
			result.S3ObjectsDeleted,
			duration.Seconds(),
			time.Now().Format(time.RFC3339),
		)
	} else {
		subject = "Video Cleanup Job Failed"
		body = fmt.Sprintf(
			"Cleanup job failed:\n\n"+
				"Error: %v\n"+
				"Videos Deleted (before failure): %d\n"+
				"S3 Objects Deleted (before failure): %d\n"+
				"Duration: %.2f seconds\n"+
				"Timestamp: %s",
			err,
			result.VideosDeleted,
			result.S3ObjectsDeleted,
			duration.Seconds(),
			time.Now().Format(time.RFC3339),
		)
	}

	event := map[string]interface{}{
		"event_type": "cleanup_job_completed",
		"user_id":    "system",
		"subject":    subject,
		"body":       body,
	}

	if notifyErr := publisher.Publish("notifications", event); notifyErr != nil {
		logger.Error("Failed to send notification", "error", notifyErr)
	} else {
		logger.Info("Notification sent", "status", status)
	}
}
