package createzip

import (
	"context"

	"github.com/video-platform/services/storage/internal/usecase/commands"
)

type CreateZipOutput struct {
	ZipPath      string
	FileCount    int
	ZipSizeBytes int64
}

type CreateZipUseCase interface {
	Execute(ctx context.Context, cmd commands.CreateZipCommand) (*CreateZipOutput, error)
}
