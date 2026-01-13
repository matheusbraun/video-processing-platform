package login

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
	"golang.org/x/crypto/bcrypt"
)

type loginUseCaseImpl struct {
	userRepo         repositories.UserRepository
	refreshTokenRepo repositories.RefreshTokenRepository
	jwtManager       jwt.JWTManager
	config           *config.Config
}

func NewLoginUseCase(
	userRepo repositories.UserRepository,
	refreshTokenRepo repositories.RefreshTokenRepository,
	jwtManager jwt.JWTManager,
	config *config.Config,
) LoginUseCase {
	return &loginUseCaseImpl{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtManager:       jwtManager,
		config:           config,
	}
}

func (uc *loginUseCaseImpl) Execute(ctx context.Context, cmd commands.LoginCommand) (*LoginOutput, error) {
	user, err := uc.userRepo.FindByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(cmd.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	accessToken, err := uc.jwtManager.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	refreshTokenString := uuid.New().String()
	refreshToken := &entities.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(uc.config.JWTRefreshExpiry),
	}

	if err := uc.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, err
	}

	return &LoginOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(uc.config.JWTAccessExpiry.Seconds()),
		UserID:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
	}, nil
}
