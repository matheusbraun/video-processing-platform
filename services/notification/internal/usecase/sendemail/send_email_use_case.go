package sendemail

import (
	"context"

	"github.com/video-platform/services/notification/internal/usecase/commands"
)

type SendEmailUseCase interface {
	Execute(ctx context.Context, cmd commands.SendEmailCommand) error
}
