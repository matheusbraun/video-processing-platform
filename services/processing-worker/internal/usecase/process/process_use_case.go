package process

import (
	"context"

	"github.com/video-platform/services/processing-worker/internal/usecase/commands"
)

type ProcessUseCase interface {
	Execute(ctx context.Context, cmd commands.ProcessCommand) error
}
