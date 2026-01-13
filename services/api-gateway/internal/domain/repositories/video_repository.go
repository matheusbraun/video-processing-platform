package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/video-platform/services/api-gateway/internal/domain/entities"
)

type VideoRepository interface {
	Create(ctx context.Context, video *entities.Video) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Video, error)
	FindByUserID(ctx context.Context, userID int64, limit, offset int) ([]*entities.Video, error)
	CountByUserID(ctx context.Context, userID int64) (int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entities.VideoStatus) error
}
