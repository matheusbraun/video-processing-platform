package logout

import (
	"context"

	"github.com/video-platform/services/auth/internal/domain/repositories"
	"github.com/video-platform/services/auth/internal/usecase/commands"
)

type logoutUseCaseImpl struct {
	refreshTokenRepo repositories.RefreshTokenRepository
}

func NewLogoutUseCase(refreshTokenRepo repositories.RefreshTokenRepository) LogoutUseCase {
	return &logoutUseCaseImpl{
		refreshTokenRepo: refreshTokenRepo,
	}
}

func (uc *logoutUseCaseImpl) Execute(ctx context.Context, cmd commands.LogoutCommand) error {
	return uc.refreshTokenRepo.DeleteByToken(ctx, cmd.RefreshToken)
}
