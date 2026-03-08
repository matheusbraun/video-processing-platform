package download

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/video-platform/services/api-gateway/internal/domain/entities"
	"github.com/video-platform/services/api-gateway/internal/domain/repositories"
	"github.com/video-platform/services/api-gateway/internal/usecase/commands"
	"github.com/video-platform/shared/pkg/storage/s3"
)

const presignedURLExpiry = 15 * time.Minute

type downloadUseCaseImpl struct {
	videoRepo           repositories.VideoRepository
	s3Client            s3.S3Client
	processedBucket     string
	s3EndpointURL       string
	s3PublicEndpointURL string
}

func NewDownloadUseCase(
	videoRepo repositories.VideoRepository,
	s3Client s3.S3Client,
	processedBucket string,
	s3EndpointURL string,
	s3PublicEndpointURL string,
) DownloadUseCase {
	return &downloadUseCaseImpl{
		videoRepo:           videoRepo,
		s3Client:            s3Client,
		processedBucket:     processedBucket,
		s3EndpointURL:       s3EndpointURL,
		s3PublicEndpointURL: s3PublicEndpointURL,
	}
}

func (uc *downloadUseCaseImpl) Execute(ctx context.Context, cmd commands.DownloadCommand) (*DownloadOutput, error) {
	video, err := uc.videoRepo.FindByID(ctx, cmd.VideoID)
	if err != nil {
		return nil, errors.New("video not found")
	}

	if video.UserID != cmd.UserID {
		return nil, errors.New("access denied")
	}

	if video.Status != entities.StatusCompleted {
		return nil, errors.New("video processing not completed")
	}

	if video.ZipPath == nil {
		return nil, errors.New("zip file not available")
	}

	presignedURL, err := uc.s3Client.GeneratePresignedURL(ctx, uc.processedBucket, *video.ZipPath, presignedURLExpiry)
	if err != nil {
		return nil, err
	}

	// Replace internal endpoint hostname with public-facing one for browser access
	if uc.s3EndpointURL != "" && uc.s3PublicEndpointURL != "" {
		presignedURL = strings.ReplaceAll(presignedURL, uc.s3EndpointURL, uc.s3PublicEndpointURL)
	}

	return &DownloadOutput{
		DownloadURL: presignedURL,
		Filename:    video.Filename + ".zip",
		ExpiresIn:   int64(presignedURLExpiry.Seconds()),
	}, nil
}
