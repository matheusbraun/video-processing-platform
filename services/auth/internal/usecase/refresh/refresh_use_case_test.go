package refresh

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/video-platform/services/auth/internal/domain/entities"
	"github.com/video-platform/services/auth/internal/usecase/commands"
	"github.com/video-platform/shared/pkg/auth/jwt"
	"github.com/video-platform/shared/pkg/config"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *entities.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id int64) (*entities.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

func (m *MockUserRepository) FindByUsername(ctx context.Context, username string) (*entities.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}

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

type MockJWTManager struct {
	mock.Mock
}

func (m *MockJWTManager) GenerateAccessToken(userID int64, email string) (string, error) {
	args := m.Called(userID, email)
	return args.String(0), args.Error(1)
}

func (m *MockJWTManager) GenerateTokenPair(userID int64, username, email string) (*jwt.TokenPair, error) {
	args := m.Called(userID, username, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.TokenPair), args.Error(1)
}

func (m *MockJWTManager) ValidateToken(token string) (*jwt.Claims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Claims), args.Error(1)
}

func TestRefreshUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockRefreshTokenRepository)
	mockJWTManager := new(MockJWTManager)

	cfg := &config.Config{
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}

	user := &entities.User{
		ID:    1,
		Email: "test@example.com",
	}

	refreshToken := &entities.RefreshToken{
		UserID:    1,
		Token:     "old_refresh_token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	cmd := commands.RefreshCommand{
		RefreshToken: "old_refresh_token",
	}

	mockTokenRepo.On("FindByToken", ctx, "old_refresh_token").Return(refreshToken, nil)
	mockUserRepo.On("FindByID", ctx, int64(1)).Return(user, nil)
	mockJWTManager.On("GenerateAccessToken", int64(1), "test@example.com").Return("new_access_token", nil)
	mockTokenRepo.On("DeleteByToken", ctx, "old_refresh_token").Return(nil)
	mockTokenRepo.On("Create", ctx, mock.AnythingOfType("*entities.RefreshToken")).Return(nil)

	useCase := NewRefreshUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, cfg)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "new_access_token", result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.NotEqual(t, "old_refresh_token", result.RefreshToken)
	assert.Equal(t, int64(900), result.ExpiresIn)

	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
	mockJWTManager.AssertExpectations(t)
}

func TestRefreshUseCase_Execute_InvalidToken(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockRefreshTokenRepository)
	mockJWTManager := new(MockJWTManager)

	cfg := &config.Config{
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}

	cmd := commands.RefreshCommand{
		RefreshToken: "invalid_token",
	}

	mockTokenRepo.On("FindByToken", ctx, "invalid_token").Return(nil, errors.New("not found"))

	useCase := NewRefreshUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, cfg)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "invalid refresh token", err.Error())

	mockTokenRepo.AssertExpectations(t)
}

func TestRefreshUseCase_Execute_ExpiredToken(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockRefreshTokenRepository)
	mockJWTManager := new(MockJWTManager)

	cfg := &config.Config{
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}

	refreshToken := &entities.RefreshToken{
		UserID:    1,
		Token:     "expired_token",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}

	cmd := commands.RefreshCommand{
		RefreshToken: "expired_token",
	}

	mockTokenRepo.On("FindByToken", ctx, "expired_token").Return(refreshToken, nil)

	useCase := NewRefreshUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, cfg)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "refresh token expired", err.Error())

	mockTokenRepo.AssertExpectations(t)
}

func TestRefreshUseCase_Execute_UserNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockRefreshTokenRepository)
	mockJWTManager := new(MockJWTManager)

	cfg := &config.Config{
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}

	refreshToken := &entities.RefreshToken{
		UserID:    999,
		Token:     "valid_token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	cmd := commands.RefreshCommand{
		RefreshToken: "valid_token",
	}

	mockTokenRepo.On("FindByToken", ctx, "valid_token").Return(refreshToken, nil)
	mockUserRepo.On("FindByID", ctx, int64(999)).Return(nil, errors.New("not found"))

	useCase := NewRefreshUseCase(mockUserRepo, mockTokenRepo, mockJWTManager, cfg)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "user not found", err.Error())

	mockTokenRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}
