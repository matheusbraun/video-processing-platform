package list

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

func TestListUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)

	videos := []*entities.Video{
		{
			ID:        uuid.New(),
			Filename:  "video1.mp4",
			Status:    entities.StatusCompleted,
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			Filename:  "video2.mp4",
			Status:    entities.StatusProcessing,
			CreatedAt: time.Now(),
		},
	}

	cmd := commands.ListCommand{
		UserID: 1,
		Limit:  10,
		Offset: 0,
	}

	mockRepo.On("FindByUserID", ctx, int64(1), 10, 0).Return(videos, nil)
	mockRepo.On("CountByUserID", ctx, int64(1)).Return(int64(2), nil)

	useCase := NewListUseCase(mockRepo)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Videos, 2)
	assert.Equal(t, int64(2), result.Total)
	assert.Equal(t, 10, result.Limit)
	assert.Equal(t, 0, result.Offset)
	assert.False(t, result.HasMore)

	mockRepo.AssertExpectations(t)
}

func TestListUseCase_Execute_EmptyList(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)

	cmd := commands.ListCommand{
		UserID: 1,
		Limit:  10,
		Offset: 0,
	}

	mockRepo.On("FindByUserID", ctx, int64(1), 10, 0).Return([]*entities.Video{}, nil)
	mockRepo.On("CountByUserID", ctx, int64(1)).Return(int64(0), nil)

	useCase := NewListUseCase(mockRepo)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Videos, 0)
	assert.Equal(t, int64(0), result.Total)
	assert.False(t, result.HasMore)

	mockRepo.AssertExpectations(t)
}

func TestListUseCase_Execute_Pagination(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)

	videos := []*entities.Video{
		{ID: uuid.New(), Filename: "video3.mp4"},
		{ID: uuid.New(), Filename: "video4.mp4"},
	}

	cmd := commands.ListCommand{
		UserID: 1,
		Limit:  2,
		Offset: 2,
	}

	mockRepo.On("FindByUserID", ctx, int64(1), 2, 2).Return(videos, nil)
	mockRepo.On("CountByUserID", ctx, int64(1)).Return(int64(5), nil)

	useCase := NewListUseCase(mockRepo)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Videos, 2)
	assert.Equal(t, int64(5), result.Total)
	assert.Equal(t, 2, result.Limit)
	assert.Equal(t, 2, result.Offset)
	assert.True(t, result.HasMore) // 2 + 2 < 5

	mockRepo.AssertExpectations(t)
}

func TestListUseCase_Execute_FindError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)

	cmd := commands.ListCommand{
		UserID: 1,
		Limit:  10,
		Offset: 0,
	}

	mockRepo.On("FindByUserID", ctx, int64(1), 10, 0).Return(nil, errors.New("database error"))

	useCase := NewListUseCase(mockRepo)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "database error", err.Error())

	mockRepo.AssertExpectations(t)
}

func TestListUseCase_Execute_CountError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockVideoRepository)

	videos := []*entities.Video{{ID: uuid.New(), Filename: "video1.mp4"}}

	cmd := commands.ListCommand{
		UserID: 1,
		Limit:  10,
		Offset: 0,
	}

	mockRepo.On("FindByUserID", ctx, int64(1), 10, 0).Return(videos, nil)
	mockRepo.On("CountByUserID", ctx, int64(1)).Return(int64(0), errors.New("count error"))

	useCase := NewListUseCase(mockRepo)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "count error", err.Error())

	mockRepo.AssertExpectations(t)
}
