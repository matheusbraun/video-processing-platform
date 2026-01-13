package controller

import (
	"context"

	"github.com/video-platform/services/processing-worker/internal/usecase/commands"
	"github.com/video-platform/services/processing-worker/internal/usecase/process"
)

type WorkerController interface {
	ProcessVideo(ctx context.Context, cmd commands.ProcessCommand) error
}

type workerControllerImpl struct {
	processUseCase process.ProcessUseCase
}

func NewWorkerController(processUseCase process.ProcessUseCase) WorkerController {
	return &workerControllerImpl{
		processUseCase: processUseCase,
	}
}

func (c *workerControllerImpl) ProcessVideo(ctx context.Context, cmd commands.ProcessCommand) error {
	return c.processUseCase.Execute(ctx, cmd)
}
