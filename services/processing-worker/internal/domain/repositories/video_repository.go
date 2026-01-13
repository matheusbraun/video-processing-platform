package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/video-platform/services/processing-worker/internal/domain/entities"
)

type VideoRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Video, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entities.VideoStatus, errorMsg *string) error
	UpdateProcessingComplete(ctx context.Context, id uuid.UUID, frameCount int, zipPath string) error
	MarkAsStarted(ctx context.Context, id uuid.UUID) error
}
