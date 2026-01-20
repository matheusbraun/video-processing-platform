package upload

import (
	"bytes"
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

func (m *MockS3Client) Upload(ctx context.Context, bucket, key string, reader io.Reader) error {
	args := m.Called(ctx, bucket, key, reader)
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

func TestUploadUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)
	mockS3 := new(MockS3Client)
	mockPublisher := new(MockPublisher)

	fileContent := []byte("fake video content")
	cmd := commands.UploadCommand{
		UserID:     1,
		Filename:   "test.mp4",
		FileSize:   int64(len(fileContent)),
		FileReader: bytes.NewReader(fileContent),
	}

	mockS3.On("Upload", ctx, "", mock.MatchedBy(func(key string) bool {
		return key != ""
	}), mock.Anything).Return(nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.Video")).Return(nil)
	mockPublisher.On("Publish", ctx, "video.processing.queue", mock.Anything).Return(nil)

	useCase := NewUploadUseCase(mockRepo, mockS3, mockPublisher)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.VideoID)
	assert.Equal(t, "test.mp4", result.Filename)
	assert.Equal(t, "PENDING", result.Status)

	mockS3.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestUploadUseCase_Execute_FileTooLarge(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)
	mockS3 := new(MockS3Client)
	mockPublisher := new(MockPublisher)

	cmd := commands.UploadCommand{
		UserID:     1,
		Filename:   "large.mp4",
		FileSize:   501 * 1024 * 1024, // 501MB
		FileReader: nil,
	}

	useCase := NewUploadUseCase(mockRepo, mockS3, mockPublisher)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "file size exceeds maximum")
}

func TestUploadUseCase_Execute_InvalidExtension(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)
	mockS3 := new(MockS3Client)
	mockPublisher := new(MockPublisher)

	cmd := commands.UploadCommand{
		UserID:     1,
		Filename:   "test.txt",
		FileSize:   1024,
		FileReader: nil,
	}

	useCase := NewUploadUseCase(mockRepo, mockS3, mockPublisher)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "file extension")
	assert.Contains(t, err.Error(), ".txt")
}

func TestUploadUseCase_Execute_S3UploadFailure(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)
	mockS3 := new(MockS3Client)
	mockPublisher := new(MockPublisher)

	fileContent := []byte("fake video content")
	cmd := commands.UploadCommand{
		UserID:     1,
		Filename:   "test.mp4",
		FileSize:   int64(len(fileContent)),
		FileReader: bytes.NewReader(fileContent),
	}

	mockS3.On("Upload", ctx, "", mock.Anything, mock.Anything).Return(errors.New("S3 error"))

	useCase := NewUploadUseCase(mockRepo, mockS3, mockPublisher)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to upload to S3")

	mockS3.AssertExpectations(t)
}

func TestUploadUseCase_Execute_DatabaseError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)
	mockS3 := new(MockS3Client)
	mockPublisher := new(MockPublisher)

	fileContent := []byte("fake video content")
	cmd := commands.UploadCommand{
		UserID:     1,
		Filename:   "test.mp4",
		FileSize:   int64(len(fileContent)),
		FileReader: bytes.NewReader(fileContent),
	}

	mockS3.On("Upload", ctx, "", mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.Video")).Return(errors.New("database error"))

	useCase := NewUploadUseCase(mockRepo, mockS3, mockPublisher)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create video record")

	mockS3.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestUploadUseCase_Execute_QueuePublishFailure(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)
	mockS3 := new(MockS3Client)
	mockPublisher := new(MockPublisher)

	fileContent := []byte("fake video content")
	cmd := commands.UploadCommand{
		UserID:     1,
		Filename:   "test.mp4",
		FileSize:   int64(len(fileContent)),
		FileReader: bytes.NewReader(fileContent),
	}

	mockS3.On("Upload", ctx, "", mock.Anything, mock.Anything).Return(nil)
	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.Video")).Return(nil)
	mockPublisher.On("Publish", ctx, "video.processing.queue", mock.Anything).Return(errors.New("queue error"))

	useCase := NewUploadUseCase(mockRepo, mockS3, mockPublisher)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to queue processing job")

	mockS3.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
}

func TestUploadUseCase_Execute_AllowedExtensions(t *testing.T) {
	allowedFiles := []string{"test.mp4", "test.avi", "test.mov", "test.mkv", "test.webm", "TEST.MP4"}

	for _, filename := range allowedFiles {
		t.Run(filename, func(t *testing.T) {
			ctx := context.Background()
			mockRepo := new(MockVideoRepository)
			mockS3 := new(MockS3Client)
			mockPublisher := new(MockPublisher)

			fileContent := []byte("fake video content")
			cmd := commands.UploadCommand{
				UserID:     1,
				Filename:   filename,
				FileSize:   int64(len(fileContent)),
				FileReader: bytes.NewReader(fileContent),
			}

			mockS3.On("Upload", ctx, "", mock.Anything, mock.Anything).Return(nil)
			mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.Video")).Return(nil)
			mockPublisher.On("Publish", ctx, "video.processing.queue", mock.Anything).Return(nil)

			useCase := NewUploadUseCase(mockRepo, mockS3, mockPublisher)

			result, err := useCase.Execute(ctx, cmd)

			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}
