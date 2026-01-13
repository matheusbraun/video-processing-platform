package download

import (
	"context"

	"github.com/video-platform/services/api-gateway/internal/usecase/commands"
)

type DownloadOutput struct {
	DownloadURL string `json:"download_url"`
	Filename    string `json:"filename"`
	ExpiresIn   int64  `json:"expires_in"`
}

type DownloadUseCase interface {
	Execute(ctx context.Context, cmd commands.DownloadCommand) (*DownloadOutput, error)
}
