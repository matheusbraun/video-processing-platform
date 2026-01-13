package list

import (
	"context"

	"github.com/video-platform/services/api-gateway/internal/domain/repositories"
	"github.com/video-platform/services/api-gateway/internal/usecase/commands"
)

type listUseCaseImpl struct {
	videoRepo repositories.VideoRepository
}

func NewListUseCase(videoRepo repositories.VideoRepository) ListUseCase {
	return &listUseCaseImpl{
		videoRepo: videoRepo,
	}
}

func (uc *listUseCaseImpl) Execute(ctx context.Context, cmd commands.ListCommand) (*ListOutput, error) {
	videos, err := uc.videoRepo.FindByUserID(ctx, cmd.UserID, cmd.Limit, cmd.Offset)
	if err != nil {
		return nil, err
	}

	total, err := uc.videoRepo.CountByUserID(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}

	videoInfos := make([]*VideoInfo, len(videos))
	for i, v := range videos {
		videoInfos[i] = &VideoInfo{
			ID:          v.ID,
			Filename:    v.Filename,
			Status:      string(v.Status),
			FrameCount:  v.FrameCount,
			CreatedAt:   v.CreatedAt,
			CompletedAt: v.CompletedAt,
		}
	}

	return &ListOutput{
		Videos:  videoInfos,
		Total:   total,
		Limit:   cmd.Limit,
		Offset:  cmd.Offset,
		HasMore: int64(cmd.Offset+cmd.Limit) < total,
	}, nil
}
