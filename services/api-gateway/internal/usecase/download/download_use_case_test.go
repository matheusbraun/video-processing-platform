package download

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/video-platform/services/api-gateway/internal/domain/entities"
	"github.com/video-platform/services/api-gateway/internal/usecase/commands"
)

type MockVideoRepository struct {
	mock.Mock
}

func (m *MockVideoRepository) Create(ctx context.Context, video *entities.Video) error {
	args := m.Called(ctx, video)
	return args.Error(0)
}

func (m *MockVideoRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Video, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Video), args.Error(1)
}

func (m *MockVideoRepository) FindByUserID(ctx context.Context, userID int64, limit, offset int) ([]*entities.Video, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Video), args.Error(1)
}

func (m *MockVideoRepository) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockVideoRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entities.VideoStatus) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

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

func TestDownloadUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)
	mockS3 := new(MockS3Client)

	videoID := uuid.New()
	zipPath := "processed/" + videoID.String() + "/frames.zip"
	video := &entities.Video{
		ID:       videoID,
		UserID:   1,
		Filename: "test.mp4",
		Status:   entities.StatusCompleted,
		ZipPath:  &zipPath,
	}

	cmd := commands.DownloadCommand{
		VideoID: videoID,
		UserID:  1,
	}

	mockRepo.On("FindByID", ctx, videoID).Return(video, nil)
	mockS3.On("GeneratePresignedURL", ctx, "", zipPath, 15*time.Minute).Return("https://s3.example.com/presigned-url", nil)

	useCase := NewDownloadUseCase(mockRepo, mockS3)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "https://s3.example.com/presigned-url", result.DownloadURL)
	assert.Equal(t, "test.mp4.zip", result.Filename)
	assert.Equal(t, int64(900), result.ExpiresIn) // 15 minutes = 900 seconds

	mockRepo.AssertExpectations(t)
	mockS3.AssertExpectations(t)
}

func TestDownloadUseCase_Execute_VideoNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)
	mockS3 := new(MockS3Client)

	videoID := uuid.New()

	cmd := commands.DownloadCommand{
		VideoID: videoID,
		UserID:  1,
	}

	mockRepo.On("FindByID", ctx, videoID).Return(nil, errors.New("not found"))

	useCase := NewDownloadUseCase(mockRepo, mockS3)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "video not found", err.Error())

	mockRepo.AssertExpectations(t)
}

func TestDownloadUseCase_Execute_AccessDenied(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)
	mockS3 := new(MockS3Client)

	videoID := uuid.New()
	zipPath := "processed/frames.zip"
	video := &entities.Video{
		ID:       videoID,
		UserID:   2, // Different user
		Filename: "test.mp4",
		Status:   entities.StatusCompleted,
		ZipPath:  &zipPath,
	}

	cmd := commands.DownloadCommand{
		VideoID: videoID,
		UserID:  1,
	}

	mockRepo.On("FindByID", ctx, videoID).Return(video, nil)

	useCase := NewDownloadUseCase(mockRepo, mockS3)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "access denied", err.Error())

	mockRepo.AssertExpectations(t)
}

func TestDownloadUseCase_Execute_ProcessingNotCompleted(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)
	mockS3 := new(MockS3Client)

	videoID := uuid.New()
	video := &entities.Video{
		ID:       videoID,
		UserID:   1,
		Filename: "test.mp4",
		Status:   entities.StatusProcessing,
	}

	cmd := commands.DownloadCommand{
		VideoID: videoID,
		UserID:  1,
	}

	mockRepo.On("FindByID", ctx, videoID).Return(video, nil)

	useCase := NewDownloadUseCase(mockRepo, mockS3)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "video processing not completed", err.Error())

	mockRepo.AssertExpectations(t)
}

func TestDownloadUseCase_Execute_ZipNotAvailable(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)
	mockS3 := new(MockS3Client)

	videoID := uuid.New()
	video := &entities.Video{
		ID:       videoID,
		UserID:   1,
		Filename: "test.mp4",
		Status:   entities.StatusCompleted,
		ZipPath:  nil,
	}

	cmd := commands.DownloadCommand{
		VideoID: videoID,
		UserID:  1,
	}

	mockRepo.On("FindByID", ctx, videoID).Return(video, nil)

	useCase := NewDownloadUseCase(mockRepo, mockS3)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "zip file not available", err.Error())

	mockRepo.AssertExpectations(t)
}

func TestDownloadUseCase_Execute_S3PresignError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)
	mockS3 := new(MockS3Client)

	videoID := uuid.New()
	zipPath := "processed/frames.zip"
	video := &entities.Video{
		ID:       videoID,
		UserID:   1,
		Filename: "test.mp4",
		Status:   entities.StatusCompleted,
		ZipPath:  &zipPath,
	}

	cmd := commands.DownloadCommand{
		VideoID: videoID,
		UserID:  1,
	}

	mockRepo.On("FindByID", ctx, videoID).Return(video, nil)
	mockS3.On("GeneratePresignedURL", ctx, "", zipPath, 15*time.Minute).Return("", errors.New("S3 error"))

	useCase := NewDownloadUseCase(mockRepo, mockS3)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "S3 error", err.Error())

	mockRepo.AssertExpectations(t)
	mockS3.AssertExpectations(t)
}
