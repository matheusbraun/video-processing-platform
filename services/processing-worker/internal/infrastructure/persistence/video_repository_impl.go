package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/video-platform/services/processing-worker/internal/domain/entities"
	"github.com/video-platform/services/processing-worker/internal/domain/repositories"
	"gorm.io/gorm"
)

type videoRepositoryImpl struct {
	db *gorm.DB
}

func NewVideoRepository(db *gorm.DB) repositories.VideoRepository {
	return &videoRepositoryImpl{db: db}
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

func (r *videoRepositoryImpl) UpdateStatus(ctx context.Context, id uuid.UUID, status entities.VideoStatus, errorMsg *string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	if errorMsg != nil {
		updates["error_message"] = *errorMsg
	}

	return r.db.WithContext(ctx).
		Model(&entities.Video{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *videoRepositoryImpl) UpdateProcessingComplete(ctx context.Context, id uuid.UUID, frameCount int, zipPath string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entities.Video{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       entities.StatusCompleted,
			"frame_count":  frameCount,
			"zip_path":     zipPath,
			"completed_at": now,
		}).Error
}

func (r *videoRepositoryImpl) MarkAsStarted(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entities.Video{}).
		Where("id = ?", id).
		Update("started_at", now).Error
}
