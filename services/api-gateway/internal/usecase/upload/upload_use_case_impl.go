package upload

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/video-platform/services/api-gateway/internal/domain/entities"
	"github.com/video-platform/services/api-gateway/internal/domain/repositories"
	"github.com/video-platform/services/api-gateway/internal/usecase/commands"
	"github.com/video-platform/shared/pkg/messaging/rabbitmq"
	"github.com/video-platform/shared/pkg/storage/s3"
)

const (
	maxFileSize = 500 * 1024 * 1024
	retention   = 15 * 24 * time.Hour
)

var allowedExtensions = map[string]bool{
	".mp4":  true,
	".avi":  true,
	".mov":  true,
	".mkv":  true,
	".webm": true,
}

type uploadUseCaseImpl struct {
	videoRepo repositories.VideoRepository
	s3Client  s3.S3Client
	publisher rabbitmq.Publisher
}

func NewUploadUseCase(
	videoRepo repositories.VideoRepository,
	s3Client s3.S3Client,
	publisher rabbitmq.Publisher,
) UploadUseCase {
	return &uploadUseCaseImpl{
		videoRepo: videoRepo,
		s3Client:  s3Client,
		publisher: publisher,
	}
}

func (uc *uploadUseCaseImpl) Execute(ctx context.Context, cmd commands.UploadCommand) (*UploadOutput, error) {
	if err := uc.validateFile(cmd); err != nil {
		return nil, err
	}

	videoID := uuid.New()
	s3Key := fmt.Sprintf("uploads/%s/%s", videoID.String(), cmd.Filename)

	if err := uc.s3Client.Upload(ctx, "", s3Key, cmd.FileReader); err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	video := &entities.Video{
		ID:           videoID,
		UserID:       cmd.UserID,
		Filename:     cmd.Filename,
		OriginalPath: s3Key,
		Status:       entities.StatusPending,
		FPS:          1,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(retention),
	}

	if err := uc.videoRepo.Create(ctx, video); err != nil {
		return nil, fmt.Errorf("failed to create video record: %w", err)
	}

	jobMessage := map[string]interface{}{
		"video_id": videoID.String(),
		"user_id":  cmd.UserID,
		"s3_key":   s3Key,
		"filename": cmd.Filename,
	}

	if err := uc.publisher.Publish(ctx, "video.processing.queue", jobMessage); err != nil {
		return nil, fmt.Errorf("failed to queue processing job: %w", err)
	}

	return &UploadOutput{
		VideoID:  videoID,
		Filename: cmd.Filename,
		Status:   string(entities.StatusPending),
	}, nil
}

func (uc *uploadUseCaseImpl) validateFile(cmd commands.UploadCommand) error {
	if cmd.FileSize > maxFileSize {
		return errors.New("file size exceeds maximum allowed (500MB)")
	}

	ext := strings.ToLower(filepath.Ext(cmd.Filename))
	if !allowedExtensions[ext] {
		return fmt.Errorf("file extension %s not allowed", ext)
	}

	return nil
}
