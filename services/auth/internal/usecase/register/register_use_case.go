package register

import (
	"context"

	"github.com/video-platform/services/auth/internal/usecase/commands"
)

type RegisterOutput struct {
	UserID   int64
	Username string
	Email    string
}

type RegisterUseCase interface {
	Execute(ctx context.Context, cmd commands.RegisterCommand) (*RegisterOutput, error)
}
