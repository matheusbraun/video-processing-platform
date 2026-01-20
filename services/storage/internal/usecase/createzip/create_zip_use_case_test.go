package createzip

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/video-platform/services/storage/internal/usecase/commands"
)

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

func TestCreateZipUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockS3 := new(MockS3Client)

	cmd := commands.CreateZipCommand{
		VideoID:   "video-123",
		S3Prefix:  "processed/video-123/frames/",
		OutputKey: "processed/video-123/frames.zip",
	}

	files := []string{
		"processed/video-123/frames/frame001.jpg",
		"processed/video-123/frames/frame002.jpg",
		"processed/video-123/frames/frame003.jpg",
	}

	mockS3.On("ListObjects", ctx, "", "processed/video-123/frames/").Return(files, nil)

	for _, file := range files {
		content := io.NopCloser(strings.NewReader("fake image data"))
		mockS3.On("GetObject", ctx, "", file).Return(content, nil)
	}

	mockS3.On("Upload", ctx, "", "processed/video-123/frames.zip", mock.Anything).Return(nil)

	useCase := NewCreateZipUseCase(mockS3)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "processed/video-123/frames.zip", result.ZipPath)
	assert.Equal(t, 3, result.FileCount)
	assert.Greater(t, result.ZipSizeBytes, int64(0))

	mockS3.AssertExpectations(t)
}

func TestCreateZipUseCase_Execute_NoFrames(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockS3 := new(MockS3Client)

	cmd := commands.CreateZipCommand{
		VideoID:   "video-123",
		S3Prefix:  "processed/video-123/frames/",
		OutputKey: "processed/video-123/frames.zip",
	}

	mockS3.On("ListObjects", ctx, "", "processed/video-123/frames/").Return([]string{}, nil)

	useCase := NewCreateZipUseCase(mockS3)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no frames found")

	mockS3.AssertExpectations(t)
}

func TestCreateZipUseCase_Execute_ListObjectsError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockS3 := new(MockS3Client)

	cmd := commands.CreateZipCommand{
		VideoID:   "video-123",
		S3Prefix:  "processed/video-123/frames/",
		OutputKey: "processed/video-123/frames.zip",
	}

	mockS3.On("ListObjects", ctx, "", "processed/video-123/frames/").Return(nil, errors.New("S3 list error"))

	useCase := NewCreateZipUseCase(mockS3)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to list frames")

	mockS3.AssertExpectations(t)
}

func TestCreateZipUseCase_Execute_GetObjectError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockS3 := new(MockS3Client)

	cmd := commands.CreateZipCommand{
		VideoID:   "video-123",
		S3Prefix:  "processed/video-123/frames/",
		OutputKey: "processed/video-123/frames.zip",
	}

	files := []string{"processed/video-123/frames/frame001.jpg"}

	mockS3.On("ListObjects", ctx, "", "processed/video-123/frames/").Return(files, nil)
	mockS3.On("GetObject", ctx, "", files[0]).Return(nil, errors.New("S3 get error"))

	useCase := NewCreateZipUseCase(mockS3)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get frame")

	mockS3.AssertExpectations(t)
}

func TestCreateZipUseCase_Execute_UploadError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockS3 := new(MockS3Client)

	cmd := commands.CreateZipCommand{
		VideoID:   "video-123",
		S3Prefix:  "processed/video-123/frames/",
		OutputKey: "processed/video-123/frames.zip",
	}

	files := []string{"processed/video-123/frames/frame001.jpg"}

	mockS3.On("ListObjects", ctx, "", "processed/video-123/frames/").Return(files, nil)
	content := io.NopCloser(strings.NewReader("fake image data"))
	mockS3.On("GetObject", ctx, "", files[0]).Return(content, nil)
	mockS3.On("Upload", ctx, "", "processed/video-123/frames.zip", mock.Anything).Return(errors.New("S3 upload error"))

	useCase := NewCreateZipUseCase(mockS3)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to upload zip")

	mockS3.AssertExpectations(t)
}

func TestCreateZipUseCase_Execute_MultipleFrames(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockS3 := new(MockS3Client)

	cmd := commands.CreateZipCommand{
		VideoID:   "video-123",
		S3Prefix:  "processed/video-123/frames/",
		OutputKey: "processed/video-123/frames.zip",
	}

	// Simulate 10 frames
	files := make([]string, 10)
	for i := 0; i < 10; i++ {
		files[i] = "processed/video-123/frames/frame" + string(rune('0'+i)) + ".jpg"
	}

	mockS3.On("ListObjects", ctx, "", "processed/video-123/frames/").Return(files, nil)

	for _, file := range files {
		content := io.NopCloser(bytes.NewReader(make([]byte, 1024))) // 1KB per frame
		mockS3.On("GetObject", ctx, "", file).Return(content, nil)
	}

	mockS3.On("Upload", ctx, "", "processed/video-123/frames.zip", mock.Anything).Return(nil)

	useCase := NewCreateZipUseCase(mockS3)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 10, result.FileCount)
	assert.Greater(t, result.ZipSizeBytes, int64(1000)) // Should be > 1KB

	mockS3.AssertExpectations(t)
}
