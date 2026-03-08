package process

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/video-platform/services/processing-worker/internal/domain/entities"
	"github.com/video-platform/services/processing-worker/internal/domain/repositories"
	"github.com/video-platform/services/processing-worker/internal/infrastructure/ffmpeg"
	"github.com/video-platform/services/processing-worker/internal/usecase/commands"
	"github.com/video-platform/shared/pkg/logging"
	"github.com/video-platform/shared/pkg/messaging/rabbitmq"
	"github.com/video-platform/shared/pkg/storage/s3"
)

type processUseCaseImpl struct {
	videoRepo       repositories.VideoRepository
	s3Client        s3.S3Client
	ffmpegService   ffmpeg.FFmpegService
	publisher       rabbitmq.Publisher
	processedBucket string
}

func NewProcessUseCase(
	videoRepo repositories.VideoRepository,
	s3Client s3.S3Client,
	ffmpegService ffmpeg.FFmpegService,
	publisher rabbitmq.Publisher,
	processedBucket string,
) ProcessUseCase {
	return &processUseCaseImpl{
		videoRepo:       videoRepo,
		s3Client:        s3Client,
		ffmpegService:   ffmpegService,
		publisher:       publisher,
		processedBucket: processedBucket,
	}
}

func (uc *processUseCaseImpl) Execute(ctx context.Context, cmd commands.ProcessCommand) error {
	logging.Info("Starting video processing", "video_id", cmd.VideoID)

	if err := uc.videoRepo.MarkAsStarted(ctx, cmd.VideoID); err != nil {
		return fmt.Errorf("failed to mark as started: %w", err)
	}

	if err := uc.videoRepo.UpdateStatus(ctx, cmd.VideoID, entities.StatusProcessing, nil); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "video-processing-*")
	if err != nil {
		return uc.handleError(ctx, cmd.VideoID, fmt.Errorf("failed to create temp dir: %w", err))
	}
	defer os.RemoveAll(tmpDir)

	videoPath := filepath.Join(tmpDir, cmd.Filename)
	framesDir := filepath.Join(tmpDir, "frames")
	if err := os.MkdirAll(framesDir, 0755); err != nil {
		return uc.handleError(ctx, cmd.VideoID, fmt.Errorf("failed to create frames dir: %w", err))
	}

	logging.Info("Downloading video from S3", "s3_key", cmd.S3Key)
	videoFile, err := os.Create(videoPath)
	if err != nil {
		return uc.handleError(ctx, cmd.VideoID, fmt.Errorf("failed to create video file: %w", err))
	}

	videoReader, err := uc.s3Client.GetObject(ctx, "", cmd.S3Key)
	if err != nil {
		videoFile.Close()
		return uc.handleError(ctx, cmd.VideoID, fmt.Errorf("failed to download video: %w", err))
	}
	defer videoReader.Close()

	if _, err := videoFile.ReadFrom(videoReader); err != nil {
		videoFile.Close()
		return uc.handleError(ctx, cmd.VideoID, fmt.Errorf("failed to write video file: %w", err))
	}
	videoFile.Close()

	logging.Info("Extracting frames with FFmpeg", "video_id", cmd.VideoID)
	frameCount, err := uc.ffmpegService.ExtractFrames(ctx, videoPath, framesDir, 1)
	if err != nil {
		return uc.handleError(ctx, cmd.VideoID, fmt.Errorf("failed to extract frames: %w", err))
	}

	logging.Info("Extracted frames", "count", frameCount)

	logging.Info("Uploading frames to S3", "video_id", cmd.VideoID)
	s3Prefix := fmt.Sprintf("processed/%s/frames/", cmd.VideoID)
	if err := uc.uploadFrames(ctx, framesDir, s3Prefix); err != nil {
		return uc.handleError(ctx, cmd.VideoID, fmt.Errorf("failed to upload frames: %w", err))
	}

	logging.Info("Creating ZIP archive", "video_id", cmd.VideoID)
	zipLocalPath := filepath.Join(tmpDir, cmd.Filename+".zip")
	if err := createZip(framesDir, zipLocalPath); err != nil {
		return uc.handleError(ctx, cmd.VideoID, fmt.Errorf("failed to create zip: %w", err))
	}

	logging.Info("Uploading ZIP to S3", "video_id", cmd.VideoID)
	zipS3Key := fmt.Sprintf("processed/%s/%s.zip", cmd.VideoID, cmd.Filename)
	zipFile, err := os.Open(zipLocalPath)
	if err != nil {
		return uc.handleError(ctx, cmd.VideoID, fmt.Errorf("failed to open zip: %w", err))
	}
	defer zipFile.Close()

	if err := uc.s3Client.Upload(ctx, uc.processedBucket, zipS3Key, zipFile); err != nil {
		return uc.handleError(ctx, cmd.VideoID, fmt.Errorf("failed to upload zip: %w", err))
	}

	zipPath := zipS3Key

	if err := uc.videoRepo.UpdateProcessingComplete(ctx, cmd.VideoID, frameCount, zipPath); err != nil {
		return uc.handleError(ctx, cmd.VideoID, fmt.Errorf("failed to update completion: %w", err))
	}

	notificationMsg := map[string]interface{}{
		"video_id":    cmd.VideoID.String(),
		"user_id":     cmd.UserID,
		"status":      "COMPLETED",
		"frame_count": frameCount,
	}

	if err := uc.publisher.Publish(ctx, "video.notification.queue", notificationMsg); err != nil {
		logging.Error("Failed to publish notification", "error", err)
	}

	logging.Info("Video processing completed", "video_id", cmd.VideoID, "frame_count", frameCount)
	return nil
}

func createZip(srcDir, destPath string) error {
	zipFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	w := zip.NewWriter(zipFile)
	defer w.Close()

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		src, err := os.Open(filepath.Join(srcDir, entry.Name()))
		if err != nil {
			return err
		}
		dst, err := w.Create(entry.Name())
		if err != nil {
			src.Close()
			return err
		}
		if _, err := io.Copy(dst, src); err != nil {
			src.Close()
			return err
		}
		src.Close()
	}
	return nil
}

func (uc *processUseCaseImpl) uploadFrames(ctx context.Context, framesDir, s3Prefix string) error {
	files, err := os.ReadDir(framesDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		framePath := filepath.Join(framesDir, file.Name())
		frameFile, err := os.Open(framePath)
		if err != nil {
			return fmt.Errorf("failed to open frame %s: %w", file.Name(), err)
		}

		s3Key := s3Prefix + file.Name()
		if err := uc.s3Client.Upload(ctx, uc.processedBucket, s3Key, frameFile); err != nil {
			frameFile.Close()
			return fmt.Errorf("failed to upload frame %s: %w", file.Name(), err)
		}
		frameFile.Close()
	}

	return nil
}

func (uc *processUseCaseImpl) handleError(ctx context.Context, videoID uuid.UUID, err error) error {
	logging.Error("Video processing failed", "video_id", videoID, "error", err)

	errMsg := err.Error()
	if updateErr := uc.videoRepo.UpdateStatus(ctx, videoID, entities.StatusFailed, &errMsg); updateErr != nil {
		logging.Error("Failed to update error status", "error", updateErr)
	}

	notificationMsg := map[string]interface{}{
		"video_id":      videoID.String(),
		"status":        "FAILED",
		"error_message": errMsg,
	}

	if pubErr := uc.publisher.Publish(ctx, "video.notification.queue", notificationMsg); pubErr != nil {
		logging.Error("Failed to publish error notification", "error", pubErr)
	}

	return err
}
