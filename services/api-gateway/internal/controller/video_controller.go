package controller

import (
	"context"

	"github.com/video-platform/services/api-gateway/internal/usecase/commands"
	"github.com/video-platform/services/api-gateway/internal/usecase/download"
	"github.com/video-platform/services/api-gateway/internal/usecase/list"
	"github.com/video-platform/services/api-gateway/internal/usecase/status"
	"github.com/video-platform/services/api-gateway/internal/usecase/upload"
)

type VideoController interface {
	Upload(ctx context.Context, cmd commands.UploadCommand) (*upload.UploadOutput, error)
	List(ctx context.Context, cmd commands.ListCommand) (*list.ListOutput, error)
	Status(ctx context.Context, cmd commands.StatusCommand) (*status.StatusOutput, error)
	Download(ctx context.Context, cmd commands.DownloadCommand) (*download.DownloadOutput, error)
}
