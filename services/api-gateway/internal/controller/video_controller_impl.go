package controller

import (
	"context"

	"github.com/video-platform/services/api-gateway/internal/usecase/commands"
	"github.com/video-platform/services/api-gateway/internal/usecase/download"
	"github.com/video-platform/services/api-gateway/internal/usecase/list"
	"github.com/video-platform/services/api-gateway/internal/usecase/status"
	"github.com/video-platform/services/api-gateway/internal/usecase/upload"
)

type videoControllerImpl struct {
	uploadUseCase   upload.UploadUseCase
	listUseCase     list.ListUseCase
	statusUseCase   status.StatusUseCase
	downloadUseCase download.DownloadUseCase
}

func NewVideoController(
	uploadUseCase upload.UploadUseCase,
	listUseCase list.ListUseCase,
	statusUseCase status.StatusUseCase,
	downloadUseCase download.DownloadUseCase,
) VideoController {
	return &videoControllerImpl{
		uploadUseCase:   uploadUseCase,
		listUseCase:     listUseCase,
		statusUseCase:   statusUseCase,
		downloadUseCase: downloadUseCase,
	}
}

func (c *videoControllerImpl) Upload(ctx context.Context, cmd commands.UploadCommand) (*upload.UploadOutput, error) {
	return c.uploadUseCase.Execute(ctx, cmd)
}

func (c *videoControllerImpl) List(ctx context.Context, cmd commands.ListCommand) (*list.ListOutput, error) {
	return c.listUseCase.Execute(ctx, cmd)
}

func (c *videoControllerImpl) Status(ctx context.Context, cmd commands.StatusCommand) (*status.StatusOutput, error) {
	return c.statusUseCase.Execute(ctx, cmd)
}

func (c *videoControllerImpl) Download(ctx context.Context, cmd commands.DownloadCommand) (*download.DownloadOutput, error) {
	return c.downloadUseCase.Execute(ctx, cmd)
}
