package status

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/video-platform/services/api-gateway/internal/usecase/commands"
)

type StatusOutput struct {
	VideoID      uuid.UUID  `json:"video_id"`
	Filename     string     `json:"filename"`
	Status       string     `json:"status"`
	FrameCount   *int       `json:"frame_count"`
	ErrorMessage *string    `json:"error_message"`
	CreatedAt    time.Time  `json:"created_at"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
}

type StatusUseCase interface {
	Execute(ctx context.Context, cmd commands.StatusCommand) (*StatusOutput, error)
}
