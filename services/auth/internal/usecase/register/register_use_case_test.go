package register

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/video-platform/services/auth/internal/domain/entities"
	"github.com/video-platform/services/auth/internal/usecase/commands"
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

func TestRegisterUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)

	cmd := commands.RegisterCommand{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "password123",
	}

	mockUserRepo.On("FindByEmail", ctx, "new@example.com").Return(nil, errors.New("not found"))
	mockUserRepo.On("FindByUsername", ctx, "newuser").Return(nil, errors.New("not found"))
	mockUserRepo.On("Create", ctx, mock.AnythingOfType("*entities.User")).Return(nil).Run(func(args mock.Arguments) {
		user := args.Get(1).(*entities.User)
		user.ID = 1 // Simulate database assigning ID
	})

	useCase := NewRegisterUseCase(mockUserRepo)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.UserID)
	assert.Equal(t, "newuser", result.Username)
	assert.Equal(t, "new@example.com", result.Email)

	mockUserRepo.AssertExpectations(t)
}

func TestRegisterUseCase_Execute_EmailAlreadyExists(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)

	cmd := commands.RegisterCommand{
		Username: "newuser",
		Email:    "existing@example.com",
		Password: "password123",
	}

	existingUser := &entities.User{
		ID:    1,
		Email: "existing@example.com",
	}

	mockUserRepo.On("FindByEmail", ctx, "existing@example.com").Return(existingUser, nil)

	useCase := NewRegisterUseCase(mockUserRepo)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "email already exists", err.Error())

	mockUserRepo.AssertExpectations(t)
}

func TestRegisterUseCase_Execute_UsernameAlreadyExists(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)

	cmd := commands.RegisterCommand{
		Username: "existinguser",
		Email:    "new@example.com",
		Password: "password123",
	}

	existingUser := &entities.User{
		ID:       1,
		Username: "existinguser",
	}

	mockUserRepo.On("FindByEmail", ctx, "new@example.com").Return(nil, errors.New("not found"))
	mockUserRepo.On("FindByUsername", ctx, "existinguser").Return(existingUser, nil)

	useCase := NewRegisterUseCase(mockUserRepo)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "username already exists", err.Error())

	mockUserRepo.AssertExpectations(t)
}

func TestRegisterUseCase_Execute_DatabaseError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)

	cmd := commands.RegisterCommand{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "password123",
	}

	mockUserRepo.On("FindByEmail", ctx, "new@example.com").Return(nil, errors.New("not found"))
	mockUserRepo.On("FindByUsername", ctx, "newuser").Return(nil, errors.New("not found"))
	mockUserRepo.On("Create", ctx, mock.AnythingOfType("*entities.User")).Return(errors.New("database error"))

	useCase := NewRegisterUseCase(mockUserRepo)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "database error", err.Error())

	mockUserRepo.AssertExpectations(t)
}

func TestRegisterUseCase_Execute_PasswordHashing(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockUserRepo := new(MockUserRepository)

	cmd := commands.RegisterCommand{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "password123",
	}

	var capturedUser *entities.User
	mockUserRepo.On("FindByEmail", ctx, "new@example.com").Return(nil, errors.New("not found"))
	mockUserRepo.On("FindByUsername", ctx, "newuser").Return(nil, errors.New("not found"))
	mockUserRepo.On("Create", ctx, mock.AnythingOfType("*entities.User")).Return(nil).Run(func(args mock.Arguments) {
		capturedUser = args.Get(1).(*entities.User)
		capturedUser.ID = 1
	})

	useCase := NewRegisterUseCase(mockUserRepo)

	// Act
	result, err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify password was hashed
	assert.NotEmpty(t, capturedUser.PasswordHash)
	assert.NotEqual(t, "password123", capturedUser.PasswordHash)

	// Verify hashed password can be verified
	err = bcrypt.CompareHashAndPassword([]byte(capturedUser.PasswordHash), []byte("password123"))
	assert.NoError(t, err)

	mockUserRepo.AssertExpectations(t)
}
