package persistence

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/video-platform/services/api-gateway/internal/domain/entities"
	"github.com/video-platform/services/api-gateway/internal/domain/repositories"
	"gorm.io/gorm"
)

type videoRepositoryImpl struct {
	db *gorm.DB
}

func NewVideoRepository(db *gorm.DB) repositories.VideoRepository {
	return &videoRepositoryImpl{db: db}
}

func (r *videoRepositoryImpl) Create(ctx context.Context, video *entities.Video) error {
	return r.db.WithContext(ctx).Create(video).Error
}

func (r *videoRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entities.Video, error) {
	var video entities.Video
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&video).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("video not found")
		}
		return nil, err
	}
	return &video, nil
}

func (r *videoRepositoryImpl) FindByUserID(ctx context.Context, userID int64, limit, offset int) ([]*entities.Video, error) {
	var videos []*entities.Video
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&videos).Error
	return videos, err
}

func (r *videoRepositoryImpl) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.Video{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

func (r *videoRepositoryImpl) UpdateStatus(ctx context.Context, id uuid.UUID, status entities.VideoStatus) error {
	return r.db.WithContext(ctx).
		Model(&entities.Video{}).
		Where("id = ?", id).
		Update("status", status).Error
}
