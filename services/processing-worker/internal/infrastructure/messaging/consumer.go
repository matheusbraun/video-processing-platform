package messaging

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/video-platform/services/processing-worker/internal/controller"
	"github.com/video-platform/services/processing-worker/internal/usecase/commands"
	"github.com/video-platform/shared/pkg/logging"
	"github.com/video-platform/shared/pkg/messaging/rabbitmq"
)

type VideoJobMessage struct {
	VideoID  string `json:"video_id"`
	UserID   int64  `json:"user_id"`
	S3Key    string `json:"s3_key"`
	Filename string `json:"filename"`
}

type VideoConsumer struct {
	consumer   *rabbitmq.Consumer
	controller controller.WorkerController
}

func NewVideoConsumer(consumer *rabbitmq.Consumer, controller controller.WorkerController) *VideoConsumer {
	return &VideoConsumer{
		consumer:   consumer,
		controller: controller,
	}
}

func (vc *VideoConsumer) Start(ctx context.Context) error {
	logging.Info("Starting video processing consumer")

	return vc.consumer.Consume(ctx, "video.processing.queue", func(body []byte) error {
		var msg VideoJobMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			logging.Error("Failed to unmarshal message", "error", err)
			return err
		}

		videoID, err := uuid.Parse(msg.VideoID)
		if err != nil {
			logging.Error("Invalid video ID", "video_id", msg.VideoID, "error", err)
			return err
		}

		cmd := commands.ProcessCommand{
			VideoID:  videoID,
			UserID:   msg.UserID,
			S3Key:    msg.S3Key,
			Filename: msg.Filename,
		}

		logging.Info("Processing video job", "video_id", videoID)

		if err := vc.controller.ProcessVideo(ctx, cmd); err != nil {
			logging.Error("Failed to process video", "video_id", videoID, "error", err)
			return err
		}

		return nil
	})
}
