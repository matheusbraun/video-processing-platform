package login

import (
	"context"

	"github.com/video-platform/services/auth/internal/usecase/commands"
)

type LoginOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	UserID       int64
	Username     string
	Email        string
}

type LoginUseCase interface {
	Execute(ctx context.Context, cmd commands.LoginCommand) (*LoginOutput, error)
}
