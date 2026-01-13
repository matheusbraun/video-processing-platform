package logout

import (
	"context"

	"github.com/video-platform/services/auth/internal/usecase/commands"
)

type LogoutUseCase interface {
	Execute(ctx context.Context, cmd commands.LogoutCommand) error
}
