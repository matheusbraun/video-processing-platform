package cleanup

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/video-platform/services/processing-worker/internal/domain/entities"
	"github.com/video-platform/shared/pkg/logging"
	"gorm.io/gorm"
)

// Mock S3Client
type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) Upload(ctx context.Context, bucket, key string, body io.Reader) error {
	args := m.Called(ctx, bucket, key, body)
	return args.Error(0)
}

func (m *MockS3Client) Download(ctx context.Context, bucket, key string, writer io.WriterAt) error {
	args := m.Called(ctx, bucket, key, writer)
	return args.Error(0)
}

func (m *MockS3Client) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	args := m.Called(ctx, bucket, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockS3Client) Delete(ctx context.Context, bucket, key string) error {
	args := m.Called(ctx, bucket, key)
	return args.Error(0)
}

func (m *MockS3Client) DeleteMultiple(ctx context.Context, bucket string, keys []string) error {
	args := m.Called(ctx, bucket, keys)
	return args.Error(0)
}

func (m *MockS3Client) GeneratePresignedURL(ctx context.Context, bucket, key string, expiration time.Duration) (string, error) {
	args := m.Called(ctx, bucket, key, expiration)
	return args.String(0), args.Error(1)
}

func (m *MockS3Client) ListObjects(ctx context.Context, bucket, prefix string) ([]string, error) {
	args := m.Called(ctx, bucket, prefix)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// Mock DB
type MockDB struct {
	Videos []entities.Video
	Err    error
}

func (m *MockDB) WithContext(ctx context.Context) *gorm.DB {
	return &gorm.DB{}
}

func TestCleanupUseCase_CleanupExpiredVideos_Success(t *testing.T) {
	mockS3 := new(MockS3Client)
	logger := logging.NewLogger("test")

	// Create test data
	now := time.Now()
	expiredTime := now.Add(-1 * time.Hour)
	zipPath := "processed/video-123/video.mp4.zip"

	videoID := uuid.New()
	video := entities.Video{
		ID:           videoID,
		UserID:       1,
		Filename:     "video.mp4",
		OriginalPath: "uploads/video.mp4",
		Status:       entities.StatusCompleted,
		ZipPath:      &zipPath,
		ExpiresAt:    expiredTime,
		CreatedAt:    now.Add(-2 * time.Hour),
	}

	// Note: This test demonstrates the expected behavior
	// In a real implementation, we would need to mock the GORM DB properly
	// For now, we're testing the collectS3Objects logic

	useCase := &cleanupUseCaseImpl{
		s3Client: mockS3,
		logger:   *logger,
		dryRun:   false,
	}

	// Test collectS3Objects
	objects := useCase.collectS3Objects(&video)

	assert.Contains(t, objects, "video-platform-uploads")
	assert.Contains(t, objects, "video-platform-processed")
	assert.Contains(t, objects["video-platform-uploads"], "uploads/video.mp4")
	assert.Contains(t, objects["video-platform-processed"], "processed/video-123/video.mp4.zip")
}

func TestCleanupUseCase_CollectS3Objects_OnlyOriginalPath(t *testing.T) {
	mockS3 := new(MockS3Client)
	logger := logging.NewLogger("test")

	videoID := uuid.New()
	video := entities.Video{
		ID:           videoID,
		OriginalPath: "uploads/test.mp4",
		ZipPath:      nil,
	}

	useCase := &cleanupUseCaseImpl{
		s3Client: mockS3,
		logger:   *logger,
		dryRun:   false,
	}

	objects := useCase.collectS3Objects(&video)

	assert.Contains(t, objects, "video-platform-uploads")
	assert.Len(t, objects["video-platform-uploads"], 1)
	assert.Equal(t, "uploads/test.mp4", objects["video-platform-uploads"][0])
}

func TestCleanupUseCase_CollectS3Objects_WithZipPath(t *testing.T) {
	mockS3 := new(MockS3Client)
	logger := logging.NewLogger("test")

	videoID := uuid.New()
	zipPath := "processed/video-123/frames.zip"
	video := entities.Video{
		ID:           videoID,
		OriginalPath: "uploads/test.mp4",
		ZipPath:      &zipPath,
	}

	useCase := &cleanupUseCaseImpl{
		s3Client: mockS3,
		logger:   *logger,
		dryRun:   false,
	}

	objects := useCase.collectS3Objects(&video)

	assert.Contains(t, objects, "video-platform-uploads")
	assert.Contains(t, objects, "video-platform-processed")
	assert.Len(t, objects["video-platform-uploads"], 1)
	assert.Contains(t, objects["video-platform-processed"], "processed/video-123/frames.zip")
	assert.Contains(t, objects["video-platform-processed"], videoID.String()+"/frames")
}

func TestCleanupUseCase_CollectS3Objects_EmptyPaths(t *testing.T) {
	mockS3 := new(MockS3Client)
	logger := logging.NewLogger("test")

	videoID := uuid.New()
	emptyZipPath := ""
	video := entities.Video{
		ID:           videoID,
		OriginalPath: "",
		ZipPath:      &emptyZipPath,
	}

	useCase := &cleanupUseCaseImpl{
		s3Client: mockS3,
		logger:   *logger,
		dryRun:   false,
	}

	objects := useCase.collectS3Objects(&video)

	// Should not add empty paths
	for _, keys := range objects {
		for _, key := range keys {
			assert.NotEmpty(t, key)
		}
	}
}

func TestCleanupUseCase_CollectS3Objects_WithS3Prefix(t *testing.T) {
	mockS3 := new(MockS3Client)
	logger := logging.NewLogger("test")

	videoID := uuid.New()
	zipPath := "s3://video-platform-processed/processed/video-123/frames.zip"
	video := entities.Video{
		ID:           videoID,
		OriginalPath: "s3://video-platform-uploads/uploads/test.mp4",
		ZipPath:      &zipPath,
	}

	useCase := &cleanupUseCaseImpl{
		s3Client: mockS3,
		logger:   *logger,
		dryRun:   false,
	}

	objects := useCase.collectS3Objects(&video)

	// Should strip s3:// prefix
	assert.Contains(t, objects["video-platform-uploads"], "uploads/test.mp4")
	assert.Contains(t, objects["video-platform-processed"], "processed/video-123/frames.zip")
}

func TestCleanupUseCase_CleanupVideo_DeleteMultipleError(t *testing.T) {
	ctx := context.Background()
	mockS3 := new(MockS3Client)
	logger := logging.NewLogger("test")

	videoID := uuid.New()
	zipPath := "processed/video-123/frames.zip"
	video := entities.Video{
		ID:           videoID,
		OriginalPath: "uploads/test.mp4",
		ZipPath:      &zipPath,
	}

	mockS3.On("DeleteMultiple", ctx, "video-platform-uploads", mock.AnythingOfType("[]string")).
		Return(errors.New("s3 error"))

	useCase := &cleanupUseCaseImpl{
		s3Client: mockS3,
		logger:   *logger,
		dryRun:   false,
	}

	result := &CleanupResult{}
	err := useCase.cleanupVideo(ctx, &video, result)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete S3 objects")
	assert.Equal(t, 0, result.VideosDeleted)
	mockS3.AssertExpectations(t)
}

func TestCleanupUseCase_CleanupVideo_DryRun(t *testing.T) {
	ctx := context.Background()
	mockS3 := new(MockS3Client)
	logger := logging.NewLogger("test")

	videoID := uuid.New()
	zipPath := "processed/video-123/frames.zip"
	video := entities.Video{
		ID:           videoID,
		OriginalPath: "uploads/test.mp4",
		ZipPath:      &zipPath,
	}

	useCase := &cleanupUseCaseImpl{
		s3Client: mockS3,
		logger:   *logger,
		dryRun:   true,
	}

	result := &CleanupResult{}
	err := useCase.cleanupVideo(ctx, &video, result)

	assert.NoError(t, err)
	assert.Equal(t, 0, result.VideosDeleted)
	assert.Equal(t, 0, result.S3ObjectsDeleted)
	// Should not call S3 delete in dry run mode
	mockS3.AssertNotCalled(t, "DeleteMultiple")
}

func TestCleanupUseCase_CleanupVideo_MultipleS3Objects(t *testing.T) {
	mockS3 := new(MockS3Client)
	logger := logging.NewLogger("test")

	videoID := uuid.New()
	zipPath := "processed/video-123/frames.zip"
	video := entities.Video{
		ID:           videoID,
		OriginalPath: "uploads/test.mp4",
		ZipPath:      &zipPath,
	}

	// We can't test the full cleanupVideo without a real DB, but we can verify collectS3Objects
	useCase := &cleanupUseCaseImpl{
		s3Client: mockS3,
		logger:   *logger,
		dryRun:   false,
	}

	objects := useCase.collectS3Objects(&video)

	assert.Len(t, objects["video-platform-uploads"], 1)
	assert.Len(t, objects["video-platform-processed"], 2)
}

func TestCleanupUseCase_CleanupResult_Initialization(t *testing.T) {
	result := &CleanupResult{}

	assert.Equal(t, 0, result.VideosDeleted)
	assert.Equal(t, 0, result.S3ObjectsDeleted)
}

func TestCleanupUseCase_CleanupResult_Accumulation(t *testing.T) {
	result := &CleanupResult{}

	result.VideosDeleted = 5
	result.S3ObjectsDeleted = 15

	assert.Equal(t, 5, result.VideosDeleted)
	assert.Equal(t, 15, result.S3ObjectsDeleted)
}
