package status

import (
	"context"
	"errors"

	"github.com/video-platform/services/api-gateway/internal/domain/repositories"
	"github.com/video-platform/services/api-gateway/internal/usecase/commands"
)

type statusUseCaseImpl struct {
	videoRepo repositories.VideoRepository
}

func NewStatusUseCase(videoRepo repositories.VideoRepository) StatusUseCase {
	return &statusUseCaseImpl{
		videoRepo: videoRepo,
	}
}

func (uc *statusUseCaseImpl) Execute(ctx context.Context, cmd commands.StatusCommand) (*StatusOutput, error) {
	video, err := uc.videoRepo.FindByID(ctx, cmd.VideoID)
	if err != nil {
		return nil, errors.New("video not found")
	}

	if video.UserID != cmd.UserID {
		return nil, errors.New("access denied")
	}

	return &StatusOutput{
		VideoID:      video.ID,
		Filename:     video.Filename,
		Status:       string(video.Status),
		FrameCount:   video.FrameCount,
		ErrorMessage: video.ErrorMessage,
		CreatedAt:    video.CreatedAt,
		StartedAt:    video.StartedAt,
		CompletedAt:  video.CompletedAt,
	}, nil
}
