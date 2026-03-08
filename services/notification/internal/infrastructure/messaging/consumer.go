package messaging

import (
	"context"
	"encoding/json"
	"time"

	"github.com/video-platform/services/notification/internal/controller"
	"github.com/video-platform/services/notification/internal/usecase/commands"
	"github.com/video-platform/shared/pkg/logging"
	"github.com/video-platform/shared/pkg/messaging/rabbitmq"
	"github.com/video-platform/shared/pkg/metrics"
)

type NotificationMessage struct {
	VideoID      string `json:"video_id"`
	UserID       int64  `json:"user_id"`
	UserEmail    string `json:"user_email"`
	Status       string `json:"status"`
	FrameCount   int    `json:"frame_count"`
	ErrorMessage string `json:"error_message"`
}

type NotificationConsumer struct {
	consumer   *rabbitmq.Consumer
	controller controller.NotificationController
	metrics    *metrics.Metrics
}

func NewNotificationConsumer(consumer *rabbitmq.Consumer, controller controller.NotificationController, m *metrics.Metrics) *NotificationConsumer {
	return &NotificationConsumer{
		consumer:   consumer,
		controller: controller,
		metrics:    m,
	}
}

func (nc *NotificationConsumer) Start(ctx context.Context) error {
	logging.Info("Starting notification consumer")

	return nc.consumer.Consume(ctx, "video.notification.queue", func(body []byte) error {
		start := time.Now()
		nc.metrics.QueueMessagesInFlight.Inc()
		defer nc.metrics.QueueMessagesInFlight.Dec()
		var msg NotificationMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			logging.Error("Failed to unmarshal message", "error", err)
			return err
		}

		logging.Info("Processing notification", "video_id", msg.VideoID, "status", msg.Status)

		if msg.UserID == 0 {
			logging.Warn("Discarding notification with invalid user_id=0, acking to prevent loop", "video_id", msg.VideoID)
			return nil
		}

		cmd := commands.SendEmailCommand{
			UserID:       msg.UserID,
			VideoID:      msg.VideoID,
			UserEmail:    msg.UserEmail,
			Status:       msg.Status,
			FrameCount:   msg.FrameCount,
			ErrorMessage: msg.ErrorMessage,
		}

		if err := nc.controller.SendEmail(ctx, cmd); err != nil {
			logging.Error("Failed to send notification, acking to prevent loop", "video_id", msg.VideoID, "error", err)
			nc.metrics.QueueMessagesProcessed.WithLabelValues("video.notification.queue", "failed").Inc()
			nc.metrics.QueueProcessingDuration.WithLabelValues("video.notification.queue").Observe(time.Since(start).Seconds())
			return nil
		}

		nc.metrics.QueueMessagesProcessed.WithLabelValues("video.notification.queue", "success").Inc()
		nc.metrics.QueueProcessingDuration.WithLabelValues("video.notification.queue").Observe(time.Since(start).Seconds())
		return nil
	})
}
