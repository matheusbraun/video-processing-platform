package login

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
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository is a mock implementation of UserRepository
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

// MockRefreshTokenRepository is a mock implementation of RefreshTokenRepository
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

// MockJWTManager is a mock implementation of JWTManager
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

func TestLoginUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockJWTManager := new(MockJWTManager)

	cfg := &config.Config{
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &entities.User{
		ID:           1,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(passwordHash),
	}

	cmd := commands.LoginCommand{
		Email:    "test@example.com",
		Password: "password123",
	}

	mockUserRepo.On("FindByEmail", ctx, "test@example.com").Return(user, nil)
	mockJWTManager.On("GenerateAccessToken", int64(1), "test@example.com").Return("access_token_123", nil)
	mockRefreshTokenRepo.On("Create", ctx, mock.AnythingOfType("*entities.RefreshToken")).Return(nil)

	useCase := NewLoginUseCase(mockUserRepo, mockRefreshTokenRepo, mockJWTManager, cfg)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "access_token_123", result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, int64(900), result.ExpiresIn) // 15 minutes = 900 seconds
	assert.Equal(t, int64(1), result.UserID)
	assert.Equal(t, "testuser", result.Username)
	assert.Equal(t, "test@example.com", result.Email)

	mockUserRepo.AssertExpectations(t)
	mockRefreshTokenRepo.AssertExpectations(t)
	mockJWTManager.AssertExpectations(t)
}

func TestLoginUseCase_Execute_UserNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockJWTManager := new(MockJWTManager)

	cfg := &config.Config{
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}

	cmd := commands.LoginCommand{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	mockUserRepo.On("FindByEmail", ctx, "nonexistent@example.com").Return(nil, errors.New("user not found"))

	useCase := NewLoginUseCase(mockUserRepo, mockRefreshTokenRepo, mockJWTManager, cfg)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "invalid credentials", err.Error())

	mockUserRepo.AssertExpectations(t)
}

func TestLoginUseCase_Execute_InvalidPassword(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockJWTManager := new(MockJWTManager)

	cfg := &config.Config{
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("correct_password"), bcrypt.DefaultCost)
	user := &entities.User{
		ID:           1,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(passwordHash),
	}

	cmd := commands.LoginCommand{
		Email:    "test@example.com",
		Password: "wrong_password",
	}

	mockUserRepo.On("FindByEmail", ctx, "test@example.com").Return(user, nil)

	useCase := NewLoginUseCase(mockUserRepo, mockRefreshTokenRepo, mockJWTManager, cfg)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "invalid credentials", err.Error())

	mockUserRepo.AssertExpectations(t)
}

func TestLoginUseCase_Execute_JWTGenerationFailure(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockJWTManager := new(MockJWTManager)

	cfg := &config.Config{
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &entities.User{
		ID:           1,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(passwordHash),
	}

	cmd := commands.LoginCommand{
		Email:    "test@example.com",
		Password: "password123",
	}

	mockUserRepo.On("FindByEmail", ctx, "test@example.com").Return(user, nil)
	mockJWTManager.On("GenerateAccessToken", int64(1), "test@example.com").Return("", errors.New("jwt generation failed"))

	useCase := NewLoginUseCase(mockUserRepo, mockRefreshTokenRepo, mockJWTManager, cfg)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "jwt generation failed", err.Error())

	mockUserRepo.AssertExpectations(t)
	mockJWTManager.AssertExpectations(t)
}

func TestLoginUseCase_Execute_RefreshTokenCreationFailure(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)
	mockRefreshTokenRepo := new(MockRefreshTokenRepository)
	mockJWTManager := new(MockJWTManager)

	cfg := &config.Config{
		JWTAccessExpiry:  15 * time.Minute,
		JWTRefreshExpiry: 7 * 24 * time.Hour,
	}

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &entities.User{
		ID:           1,
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(passwordHash),
	}

	cmd := commands.LoginCommand{
		Email:    "test@example.com",
		Password: "password123",
	}

	mockUserRepo.On("FindByEmail", ctx, "test@example.com").Return(user, nil)
	mockJWTManager.On("GenerateAccessToken", int64(1), "test@example.com").Return("access_token_123", nil)
	mockRefreshTokenRepo.On("Create", ctx, mock.AnythingOfType("*entities.RefreshToken")).Return(errors.New("database error"))

	useCase := NewLoginUseCase(mockUserRepo, mockRefreshTokenRepo, mockJWTManager, cfg)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "database error", err.Error())

	mockUserRepo.AssertExpectations(t)
	mockJWTManager.AssertExpectations(t)
	mockRefreshTokenRepo.AssertExpectations(t)
}
