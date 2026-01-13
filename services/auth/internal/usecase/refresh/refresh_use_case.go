package refresh

import (
	"context"

	"github.com/video-platform/services/auth/internal/usecase/commands"
)

type RefreshOutput struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type RefreshUseCase interface {
	Execute(ctx context.Context, cmd commands.RefreshCommand) (*RefreshOutput, error)
}
