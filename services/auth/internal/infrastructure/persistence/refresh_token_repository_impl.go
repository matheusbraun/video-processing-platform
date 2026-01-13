package persistence

import (
	"context"
	"errors"

	"github.com/video-platform/services/auth/internal/domain/entities"
	"github.com/video-platform/services/auth/internal/domain/repositories"
	"gorm.io/gorm"
)

type refreshTokenRepositoryImpl struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) repositories.RefreshTokenRepository {
	return &refreshTokenRepositoryImpl{db: db}
}

func (r *refreshTokenRepositoryImpl) Create(ctx context.Context, token *entities.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *refreshTokenRepositoryImpl) FindByToken(ctx context.Context, token string) (*entities.RefreshToken, error) {
	var refreshToken entities.RefreshToken
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&refreshToken).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("token not found")
		}
		return nil, err
	}
	return &refreshToken, nil
}

func (r *refreshTokenRepositoryImpl) DeleteByToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&entities.RefreshToken{}).Error
}

func (r *refreshTokenRepositoryImpl) DeleteByUserID(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&entities.RefreshToken{}).Error
}
