package repositories

import (
	"context"

	"github.com/video-platform/services/auth/internal/domain/entities"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entities.RefreshToken) error
	FindByToken(ctx context.Context, token string) (*entities.RefreshToken, error)
	DeleteByToken(ctx context.Context, token string) error
	DeleteByUserID(ctx context.Context, userID int64) error
}
