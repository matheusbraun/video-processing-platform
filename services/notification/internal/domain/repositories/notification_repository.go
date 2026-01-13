package repositories

import (
	"context"

	"github.com/video-platform/services/notification/internal/domain/entities"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *entities.Notification) error
	MarkAsSent(ctx context.Context, id int64) error
	MarkAsFailed(ctx context.Context, id int64, errorMsg string) error
}
