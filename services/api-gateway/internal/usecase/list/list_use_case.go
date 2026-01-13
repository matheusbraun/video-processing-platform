package list

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/video-platform/services/api-gateway/internal/usecase/commands"
)

type VideoInfo struct {
	ID          uuid.UUID  `json:"id"`
	Filename    string     `json:"filename"`
	Status      string     `json:"status"`
	FrameCount  *int       `json:"frame_count"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

type ListOutput struct {
	Videos  []*VideoInfo `json:"videos"`
	Total   int64        `json:"total"`
	Limit   int          `json:"limit"`
	Offset  int          `json:"offset"`
	HasMore bool         `json:"has_more"`
}

type ListUseCase interface {
	Execute(ctx context.Context, cmd commands.ListCommand) (*ListOutput, error)
}
