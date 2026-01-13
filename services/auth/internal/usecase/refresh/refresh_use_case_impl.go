package refresh

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/video-platform/services/auth/internal/domain/entities"
	"github.com/video-platform/services/auth/internal/domain/repositories"
	"github.com/video-platform/services/auth/internal/usecase/commands"
	"github.com/video-platform/shared/pkg/auth/jwt"
	"github.com/video-platform/shared/pkg/config"
)

type refreshUseCaseImpl struct {
	userRepo         repositories.UserRepository
	refreshTokenRepo repositories.RefreshTokenRepository
	jwtManager       jwt.JWTManager
	config           *config.Config
}

func NewRefreshUseCase(
	userRepo repositories.UserRepository,
	refreshTokenRepo repositories.RefreshTokenRepository,
	jwtManager jwt.JWTManager,
	config *config.Config,
) RefreshUseCase {
	return &refreshUseCaseImpl{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtManager:       jwtManager,
		config:           config,
	}
}

func (uc *refreshUseCaseImpl) Execute(ctx context.Context, cmd commands.RefreshCommand) (*RefreshOutput, error) {
	token, err := uc.refreshTokenRepo.FindByToken(ctx, cmd.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	if token.IsExpired() {
		return nil, errors.New("refresh token expired")
	}

	user, err := uc.userRepo.FindByID(ctx, token.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	accessToken, err := uc.jwtManager.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	if err := uc.refreshTokenRepo.DeleteByToken(ctx, cmd.RefreshToken); err != nil {
		return nil, err
	}

	newRefreshTokenString := uuid.New().String()
	newRefreshToken := &entities.RefreshToken{
		UserID:    user.ID,
		Token:     newRefreshTokenString,
		ExpiresAt: time.Now().Add(uc.config.JWTRefreshExpiry),
	}

	if err := uc.refreshTokenRepo.Create(ctx, newRefreshToken); err != nil {
		return nil, err
	}

	return &RefreshOutput{
		AccessToken:  accessToken,
		RefreshToken: newRefreshTokenString,
		ExpiresIn:    int64(uc.config.JWTAccessExpiry.Seconds()),
	}, nil
}
