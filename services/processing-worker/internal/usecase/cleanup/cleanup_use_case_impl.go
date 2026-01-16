package cleanup

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/video-platform/services/processing-worker/internal/domain/entities"
	"github.com/video-platform/shared/pkg/logging"
	"github.com/video-platform/shared/pkg/storage/s3"
	"gorm.io/gorm"
)

type cleanupUseCaseImpl struct {
	db       *gorm.DB
	s3Client s3.S3Client
	logger   logging.Logger
	dryRun   bool
}

func NewCleanupUseCaseImpl(
	db *gorm.DB,
	s3Client s3.S3Client,
	logger logging.Logger,
	dryRun bool,
) CleanupUseCase {
	return &cleanupUseCaseImpl{
		db:       db,
		s3Client: s3Client,
		logger:   logger,
		dryRun:   dryRun,
	}
}

func (uc *cleanupUseCaseImpl) CleanupExpiredVideos(ctx context.Context) (*CleanupResult, error) {
	result := &CleanupResult{}

	var expiredVideos []entities.Video
	err := uc.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Find(&expiredVideos).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query expired videos: %w", err)
	}

	uc.logger.Info("Found expired videos", "count", len(expiredVideos))

	for _, video := range expiredVideos {
		if err := uc.cleanupVideo(ctx, &video, result); err != nil {
			uc.logger.Error("Failed to cleanup video", "video_id", video.ID, "error", err)
			continue
		}
	}

	return result, nil
}

func (uc *cleanupUseCaseImpl) cleanupVideo(ctx context.Context, video *entities.Video, result *CleanupResult) error {
	s3Objects := uc.collectS3Objects(video)

	if !uc.dryRun {
		for bucket, keys := range s3Objects {
			if len(keys) == 0 {
				continue
			}

			uc.logger.Info("Deleting S3 objects", "bucket", bucket, "count", len(keys), "video_id", video.ID)

			if err := uc.s3Client.DeleteMultiple(ctx, bucket, keys); err != nil {
				return fmt.Errorf("failed to delete S3 objects: %w", err)
			}

			result.S3ObjectsDeleted += len(keys)
		}

		if err := uc.db.WithContext(ctx).Delete(video).Error; err != nil {
			return fmt.Errorf("failed to delete video from database: %w", err)
		}

		result.VideosDeleted++
		uc.logger.Info("Video cleanup completed", "video_id", video.ID)
	} else {
		totalObjects := 0
		for _, keys := range s3Objects {
			totalObjects += len(keys)
		}
		uc.logger.Info("DRY RUN: Would delete video",
			"video_id", video.ID,
			"s3_objects", totalObjects,
		)
	}

	return nil
}

func (uc *cleanupUseCaseImpl) collectS3Objects(video *entities.Video) map[string][]string {
	objects := make(map[string][]string)

	uploadsBucket := "video-platform-uploads"
	processedBucket := "video-platform-processed"

	if video.OriginalPath != "" {
		key := strings.TrimPrefix(video.OriginalPath, "s3://"+uploadsBucket+"/")
		objects[uploadsBucket] = append(objects[uploadsBucket], key)
	}

	if video.ZipPath != nil && *video.ZipPath != "" {
		zipKey := strings.TrimPrefix(*video.ZipPath, "s3://"+processedBucket+"/")
		objects[processedBucket] = append(objects[processedBucket], zipKey)

		framePrefix := filepath.Join(video.ID.String(), "frames/")
		objects[processedBucket] = append(objects[processedBucket], framePrefix)
	}

	return objects
}
