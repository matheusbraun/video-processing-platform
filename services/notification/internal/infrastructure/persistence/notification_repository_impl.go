package persistence

import (
	"context"
	"time"

	"github.com/video-platform/services/notification/internal/domain/entities"
	"github.com/video-platform/services/notification/internal/domain/repositories"
	"gorm.io/gorm"
)

type notificationRepositoryImpl struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) repositories.NotificationRepository {
	return &notificationRepositoryImpl{db: db}
}

func (r *notificationRepositoryImpl) Create(ctx context.Context, notification *entities.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *notificationRepositoryImpl) MarkAsSent(ctx context.Context, id int64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&entities.Notification{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":  entities.StatusSent,
			"sent_at": now,
		}).Error
}

func (r *notificationRepositoryImpl) MarkAsFailed(ctx context.Context, id int64, errorMsg string) error {
	return r.db.WithContext(ctx).
		Model(&entities.Notification{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":        entities.StatusFailed,
			"error_message": errorMsg,
		}).Error
}
