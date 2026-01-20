package status

import (
	"context"
	"errors"
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

func TestStatusUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)

	videoID := uuid.New()
	createdAt := time.Now()
	frameCount := 100
	video := &entities.Video{
		ID:         videoID,
		UserID:     1,
		Filename:   "test.mp4",
		Status:     entities.StatusCompleted,
		FrameCount: &frameCount,
		CreatedAt:  createdAt,
	}

	cmd := commands.StatusCommand{
		VideoID: videoID,
		UserID:  1,
	}

	mockRepo.On("FindByID", ctx, videoID).Return(video, nil)

	useCase := NewStatusUseCase(mockRepo)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, videoID, result.VideoID)
	assert.Equal(t, "test.mp4", result.Filename)
	assert.Equal(t, "COMPLETED", result.Status)
	assert.NotNil(t, result.FrameCount)
	assert.Equal(t, 100, *result.FrameCount)

	mockRepo.AssertExpectations(t)
}

func TestStatusUseCase_Execute_VideoNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)

	videoID := uuid.New()

	cmd := commands.StatusCommand{
		VideoID: videoID,
		UserID:  1,
	}

	mockRepo.On("FindByID", ctx, videoID).Return(nil, errors.New("not found"))

	useCase := NewStatusUseCase(mockRepo)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "video not found", err.Error())

	mockRepo.AssertExpectations(t)
}

func TestStatusUseCase_Execute_AccessDenied(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)

	videoID := uuid.New()
	video := &entities.Video{
		ID:       videoID,
		UserID:   2, // Different user
		Filename: "test.mp4",
		Status:   entities.StatusCompleted,
	}

	cmd := commands.StatusCommand{
		VideoID: videoID,
		UserID:  1,
	}

	mockRepo.On("FindByID", ctx, videoID).Return(video, nil)

	useCase := NewStatusUseCase(mockRepo)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "access denied", err.Error())

	mockRepo.AssertExpectations(t)
}

func TestStatusUseCase_Execute_ProcessingStatus(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)

	videoID := uuid.New()
	startedAt := time.Now()
	video := &entities.Video{
		ID:        videoID,
		UserID:    1,
		Filename:  "test.mp4",
		Status:    entities.StatusProcessing,
		StartedAt: &startedAt,
		CreatedAt: time.Now().Add(-5 * time.Minute),
	}

	cmd := commands.StatusCommand{
		VideoID: videoID,
		UserID:  1,
	}

	mockRepo.On("FindByID", ctx, videoID).Return(video, nil)

	useCase := NewStatusUseCase(mockRepo)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "PROCESSING", result.Status)
	assert.NotNil(t, result.StartedAt)

	mockRepo.AssertExpectations(t)
}

func TestStatusUseCase_Execute_FailedStatus(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)

	videoID := uuid.New()
	errorMsg := "FFmpeg processing failed"
	video := &entities.Video{
		ID:           videoID,
		UserID:       1,
		Filename:     "test.mp4",
		Status:       entities.StatusFailed,
		ErrorMessage: &errorMsg,
		CreatedAt:    time.Now(),
	}

	cmd := commands.StatusCommand{
		VideoID: videoID,
		UserID:  1,
	}

	mockRepo.On("FindByID", ctx, videoID).Return(video, nil)

	useCase := NewStatusUseCase(mockRepo)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "FAILED", result.Status)
	assert.NotNil(t, result.ErrorMessage)
	assert.Equal(t, "FFmpeg processing failed", *result.ErrorMessage)

	mockRepo.AssertExpectations(t)
}
