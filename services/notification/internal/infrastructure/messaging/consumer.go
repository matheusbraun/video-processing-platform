package messaging

import (
	"context"
	"encoding/json"

	"github.com/video-platform/services/notification/internal/controller"
	"github.com/video-platform/services/notification/internal/usecase/commands"
	"github.com/video-platform/shared/pkg/logging"
	"github.com/video-platform/shared/pkg/messaging/rabbitmq"
)

type NotificationMessage struct {
	VideoID      string `json:"video_id"`
	UserID       int64  `json:"user_id"`
	Status       string `json:"status"`
	FrameCount   int    `json:"frame_count"`
	ErrorMessage string `json:"error_message"`
}

type NotificationConsumer struct {
	consumer   *rabbitmq.Consumer
	controller controller.NotificationController
}

func NewNotificationConsumer(consumer *rabbitmq.Consumer, controller controller.NotificationController) *NotificationConsumer {
	return &NotificationConsumer{
		consumer:   consumer,
		controller: controller,
	}
}

func (nc *NotificationConsumer) Start(ctx context.Context) error {
	logging.Info("Starting notification consumer")

	return nc.consumer.Consume(ctx, "video.notification.queue", func(body []byte) error {
		var msg NotificationMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			logging.Error("Failed to unmarshal message", "error", err)
			return err
		}

		logging.Info("Processing notification", "video_id", msg.VideoID, "status", msg.Status)

		cmd := commands.SendEmailCommand{
			UserID:       msg.UserID,
			VideoID:      msg.VideoID,
			UserEmail:    "user@example.com",
			Status:       msg.Status,
			FrameCount:   msg.FrameCount,
			ErrorMessage: msg.ErrorMessage,
		}

		if err := nc.controller.SendEmail(ctx, cmd); err != nil {
			logging.Error("Failed to send notification", "video_id", msg.VideoID, "error", err)
			return err
		}

		return nil
	})
}
