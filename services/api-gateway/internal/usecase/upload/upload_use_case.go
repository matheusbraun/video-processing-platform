package upload

import (
	"context"

	"github.com/google/uuid"
	"github.com/video-platform/services/api-gateway/internal/usecase/commands"
)

type UploadOutput struct {
	VideoID  uuid.UUID
	Filename string
	Status   string
}

type UploadUseCase interface {
	Execute(ctx context.Context, cmd commands.UploadCommand) (*UploadOutput, error)
}
