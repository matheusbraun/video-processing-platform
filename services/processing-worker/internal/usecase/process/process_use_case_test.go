package process

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/video-platform/services/processing-worker/internal/domain/entities"
	"github.com/video-platform/services/processing-worker/internal/usecase/commands"
)

// Mock VideoRepository
type MockVideoRepository struct {
	mock.Mock
}

func (m *MockVideoRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Video, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Video), args.Error(1)
}

func (m *MockVideoRepository) MarkAsStarted(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockVideoRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entities.VideoStatus, errorMessage *string) error {
	args := m.Called(ctx, id, status, errorMessage)
	return args.Error(0)
}

func (m *MockVideoRepository) UpdateProcessingComplete(ctx context.Context, id uuid.UUID, frameCount int, zipPath string) error {
	args := m.Called(ctx, id, frameCount, zipPath)
	return args.Error(0)
}

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

// Mock FFmpegService
type MockFFmpegService struct {
	mock.Mock
}

func (m *MockFFmpegService) ExtractFrames(ctx context.Context, videoPath, outputDir string, fps int) (int, error) {
	args := m.Called(ctx, videoPath, outputDir, fps)
	return args.Int(0), args.Error(1)
}

// Mock Publisher
type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(ctx context.Context, queue string, message interface{}) error {
	args := m.Called(ctx, queue, message)
	return args.Error(0)
}

func (m *MockPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestProcessUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)
	mockS3 := new(MockS3Client)
	mockFFmpeg := new(MockFFmpegService)
	mockPublisher := new(MockPublisher)

	videoID := uuid.New()
	cmd := commands.ProcessCommand{
		VideoID:  videoID,
		UserID:   1,
		S3Key:    "uploads/video.mp4",
		Filename: "video.mp4",
	}

	// Setup expectations
	mockRepo.On("MarkAsStarted", ctx, videoID).Return(nil)
	mockRepo.On("UpdateStatus", ctx, videoID, entities.StatusProcessing, (*string)(nil)).Return(nil)

	// Mock S3 download
	videoContent := io.NopCloser(strings.NewReader("fake video content"))
	mockS3.On("GetObject", ctx, "", "uploads/video.mp4").Return(videoContent, nil)

	// Mock FFmpeg frame extraction
	mockFFmpeg.On("ExtractFrames", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 1).Return(10, nil)

	// Mock S3 frame uploads (may be called if FFmpeg creates actual files, but in unit tests it won't)
	mockS3.On("Upload", ctx, "processed-bucket", mock.AnythingOfType("string"), mock.Anything).Return(nil).Maybe()

	// Mock repository update for completion
	mockRepo.On("UpdateProcessingComplete", ctx, videoID, 10, mock.MatchedBy(func(path string) bool {
		return strings.Contains(path, "video.mp4.zip")
	})).Return(nil)

	// Mock notification publishing
	mockPublisher.On("Publish", ctx, "video.notification.queue", mock.MatchedBy(func(msg interface{}) bool {
		m := msg.(map[string]interface{})
		return m["video_id"] == videoID.String() && m["status"] == "COMPLETED"
	})).Return(nil)

	useCase := NewProcessUseCase(mockRepo, mockS3, mockFFmpeg, mockPublisher, "processed-bucket")
	err := useCase.Execute(ctx, cmd)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockS3.AssertExpectations(t)
	mockFFmpeg.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestProcessUseCase_Execute_MarkAsStartedError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)
	mockS3 := new(MockS3Client)
	mockFFmpeg := new(MockFFmpegService)
	mockPublisher := new(MockPublisher)

	videoID := uuid.New()
	cmd := commands.ProcessCommand{
		VideoID:  videoID,
		UserID:   1,
		S3Key:    "uploads/video.mp4",
		Filename: "video.mp4",
	}

	mockRepo.On("MarkAsStarted", ctx, videoID).Return(errors.New("database error"))

	useCase := NewProcessUseCase(mockRepo, mockS3, mockFFmpeg, mockPublisher, "processed-bucket")
	err := useCase.Execute(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to mark as started")
	mockRepo.AssertExpectations(t)
}

func TestProcessUseCase_Execute_UpdateStatusError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)
	mockS3 := new(MockS3Client)
	mockFFmpeg := new(MockFFmpegService)
	mockPublisher := new(MockPublisher)

	videoID := uuid.New()
	cmd := commands.ProcessCommand{
		VideoID:  videoID,
		UserID:   1,
		S3Key:    "uploads/video.mp4",
		Filename: "video.mp4",
	}

	mockRepo.On("MarkAsStarted", ctx, videoID).Return(nil)
	mockRepo.On("UpdateStatus", ctx, videoID, entities.StatusProcessing, (*string)(nil)).Return(errors.New("database error"))

	useCase := NewProcessUseCase(mockRepo, mockS3, mockFFmpeg, mockPublisher, "processed-bucket")
	err := useCase.Execute(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update status")
	mockRepo.AssertExpectations(t)
}

func TestProcessUseCase_Execute_DownloadVideoError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)
	mockS3 := new(MockS3Client)
	mockFFmpeg := new(MockFFmpegService)
	mockPublisher := new(MockPublisher)

	videoID := uuid.New()
	cmd := commands.ProcessCommand{
		VideoID:  videoID,
		UserID:   1,
		S3Key:    "uploads/video.mp4",
		Filename: "video.mp4",
	}

	mockRepo.On("MarkAsStarted", ctx, videoID).Return(nil)
	mockRepo.On("UpdateStatus", ctx, videoID, entities.StatusProcessing, (*string)(nil)).Return(nil)
	mockS3.On("GetObject", ctx, "", "uploads/video.mp4").Return(nil, errors.New("s3 error"))

	// Expect error handling
	mockRepo.On("UpdateStatus", ctx, videoID, entities.StatusFailed, mock.AnythingOfType("*string")).Return(nil)
	mockPublisher.On("Publish", ctx, "video.notification.queue", mock.MatchedBy(func(msg interface{}) bool {
		m := msg.(map[string]interface{})
		return m["status"] == "FAILED"
	})).Return(nil)

	useCase := NewProcessUseCase(mockRepo, mockS3, mockFFmpeg, mockPublisher, "processed-bucket")
	err := useCase.Execute(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to download video")
	mockRepo.AssertExpectations(t)
	mockS3.AssertExpectations(t)
}

func TestProcessUseCase_Execute_FFmpegError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)
	mockS3 := new(MockS3Client)
	mockFFmpeg := new(MockFFmpegService)
	mockPublisher := new(MockPublisher)

	videoID := uuid.New()
	cmd := commands.ProcessCommand{
		VideoID:  videoID,
		UserID:   1,
		S3Key:    "uploads/video.mp4",
		Filename: "video.mp4",
	}

	mockRepo.On("MarkAsStarted", ctx, videoID).Return(nil)
	mockRepo.On("UpdateStatus", ctx, videoID, entities.StatusProcessing, (*string)(nil)).Return(nil)

	videoContent := io.NopCloser(strings.NewReader("fake video content"))
	mockS3.On("GetObject", ctx, "", "uploads/video.mp4").Return(videoContent, nil)

	mockFFmpeg.On("ExtractFrames", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 1).Return(0, errors.New("ffmpeg error"))

	// Expect error handling
	mockRepo.On("UpdateStatus", ctx, videoID, entities.StatusFailed, mock.AnythingOfType("*string")).Return(nil)
	mockPublisher.On("Publish", ctx, "video.notification.queue", mock.MatchedBy(func(msg interface{}) bool {
		m := msg.(map[string]interface{})
		return m["status"] == "FAILED"
	})).Return(nil)

	useCase := NewProcessUseCase(mockRepo, mockS3, mockFFmpeg, mockPublisher, "processed-bucket")
	err := useCase.Execute(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to extract frames")
	mockRepo.AssertExpectations(t)
	mockFFmpeg.AssertExpectations(t)
}

func TestProcessUseCase_Execute_UploadFramesError(t *testing.T) {
	// Note: This test cannot properly test upload failures because uploadFrames reads actual files from disk.
	// Since FFmpeg is mocked and doesn't create real files, uploadFrames succeeds (empty directory).
	// To properly test upload failures, we would need to either:
	// 1. Create actual frame files in the temp directory
	// 2. Refactor uploadFrames to be more testable (inject file reader)
	// For now, we skip this test as it would require significant refactoring.
	t.Skip("Skipping: uploadFrames cannot be properly tested without filesystem integration")
}

func TestProcessUseCase_Execute_UpdateCompletionError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)
	mockS3 := new(MockS3Client)
	mockFFmpeg := new(MockFFmpegService)
	mockPublisher := new(MockPublisher)

	videoID := uuid.New()
	cmd := commands.ProcessCommand{
		VideoID:  videoID,
		UserID:   1,
		S3Key:    "uploads/video.mp4",
		Filename: "video.mp4",
	}

	mockRepo.On("MarkAsStarted", ctx, videoID).Return(nil)
	mockRepo.On("UpdateStatus", ctx, videoID, entities.StatusProcessing, (*string)(nil)).Return(nil)

	videoContent := io.NopCloser(strings.NewReader("fake video content"))
	mockS3.On("GetObject", ctx, "", "uploads/video.mp4").Return(videoContent, nil)

	mockFFmpeg.On("ExtractFrames", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 1).Return(10, nil)
	mockS3.On("Upload", ctx, "processed-bucket", mock.AnythingOfType("string"), mock.Anything).Return(nil)

	mockRepo.On("UpdateProcessingComplete", ctx, videoID, 10, mock.AnythingOfType("string")).Return(errors.New("database error"))

	// Expect error handling
	mockRepo.On("UpdateStatus", ctx, videoID, entities.StatusFailed, mock.AnythingOfType("*string")).Return(nil)
	mockPublisher.On("Publish", ctx, "video.notification.queue", mock.MatchedBy(func(msg interface{}) bool {
		m := msg.(map[string]interface{})
		return m["status"] == "FAILED"
	})).Return(nil)

	useCase := NewProcessUseCase(mockRepo, mockS3, mockFFmpeg, mockPublisher, "processed-bucket")
	err := useCase.Execute(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update completion")
	mockRepo.AssertExpectations(t)
}

func TestProcessUseCase_Execute_NotificationPublishError(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)
	mockS3 := new(MockS3Client)
	mockFFmpeg := new(MockFFmpegService)
	mockPublisher := new(MockPublisher)

	videoID := uuid.New()
	cmd := commands.ProcessCommand{
		VideoID:  videoID,
		UserID:   1,
		S3Key:    "uploads/video.mp4",
		Filename: "video.mp4",
	}

	mockRepo.On("MarkAsStarted", ctx, videoID).Return(nil)
	mockRepo.On("UpdateStatus", ctx, videoID, entities.StatusProcessing, (*string)(nil)).Return(nil)

	videoContent := io.NopCloser(strings.NewReader("fake video content"))
	mockS3.On("GetObject", ctx, "", "uploads/video.mp4").Return(videoContent, nil)

	mockFFmpeg.On("ExtractFrames", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), 1).Return(10, nil)
	mockS3.On("Upload", ctx, "processed-bucket", mock.AnythingOfType("string"), mock.Anything).Return(nil)
	mockRepo.On("UpdateProcessingComplete", ctx, videoID, 10, mock.AnythingOfType("string")).Return(nil)

	// Notification publish fails, but should not fail the use case
	mockPublisher.On("Publish", ctx, "video.notification.queue", mock.Anything).Return(errors.New("rabbitmq error"))

	useCase := NewProcessUseCase(mockRepo, mockS3, mockFFmpeg, mockPublisher, "processed-bucket")
	err := useCase.Execute(ctx, cmd)

	// Should still succeed even if notification fails
	assert.NoError(t, err)
	mockPublisher.AssertExpectations(t)
}
