package controller

import (
	"context"

	"github.com/video-platform/services/storage/internal/usecase/commands"
	"github.com/video-platform/services/storage/internal/usecase/createzip"
)

type StorageController interface {
	CreateZip(ctx context.Context, cmd commands.CreateZipCommand) (*createzip.CreateZipOutput, error)
}

type storageControllerImpl struct {
	createZipUseCase createzip.CreateZipUseCase
}

func NewStorageController(createZipUseCase createzip.CreateZipUseCase) StorageController {
	return &storageControllerImpl{
		createZipUseCase: createZipUseCase,
	}
}

func (c *storageControllerImpl) CreateZip(ctx context.Context, cmd commands.CreateZipCommand) (*createzip.CreateZipOutput, error) {
	return c.createZipUseCase.Execute(ctx, cmd)
}
