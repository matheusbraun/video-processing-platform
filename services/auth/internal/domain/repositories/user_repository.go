package repositories

import (
	"context"

	"github.com/video-platform/services/auth/internal/domain/entities"
)

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	FindByUsername(ctx context.Context, username string) (*entities.User, error)
	FindByID(ctx context.Context, id int64) (*entities.User, error)
}
