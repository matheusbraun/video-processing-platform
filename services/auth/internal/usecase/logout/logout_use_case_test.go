package logout

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/video-platform/services/auth/internal/domain/entities"
	"github.com/video-platform/services/auth/internal/usecase/commands"
)

type MockRefreshTokenRepository struct {
	mock.Mock
}

func (m *MockRefreshTokenRepository) Create(ctx context.Context, token *entities.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) FindByToken(ctx context.Context, token string) (*entities.RefreshToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.RefreshToken), args.Error(1)
}

func (m *MockRefreshTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockRefreshTokenRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestLogoutUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockRefreshTokenRepository)

	cmd := commands.LogoutCommand{
		RefreshToken: "valid_token_123",
	}

	mockRepo.On("DeleteByToken", ctx, "valid_token_123").Return(nil)

	useCase := NewLogoutUseCase(mockRepo)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestLogoutUseCase_Execute_TokenNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockRefreshTokenRepository)

	cmd := commands.LogoutCommand{
		RefreshToken: "nonexistent_token",
	}

	mockRepo.On("DeleteByToken", ctx, "nonexistent_token").Return(errors.New("token not found"))

	useCase := NewLogoutUseCase(mockRepo)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "token not found", err.Error())
	mockRepo.AssertExpectations(t)
}

func TestLogoutUseCase_Execute_DatabaseError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockRefreshTokenRepository)

	cmd := commands.LogoutCommand{
		RefreshToken: "valid_token_123",
	}

	mockRepo.On("DeleteByToken", ctx, "valid_token_123").Return(errors.New("database error"))

	useCase := NewLogoutUseCase(mockRepo)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	mockRepo.AssertExpectations(t)
}
