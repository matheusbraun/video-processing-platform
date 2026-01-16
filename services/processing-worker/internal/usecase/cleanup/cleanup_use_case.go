package cleanup

import "context"

type CleanupResult struct {
	VideosDeleted    int
	S3ObjectsDeleted int
}

type CleanupUseCase interface {
	CleanupExpiredVideos(ctx context.Context) (*CleanupResult, error)
}
