package register

import (
	"context"
	"errors"

	"github.com/video-platform/services/auth/internal/domain/entities"
	"github.com/video-platform/services/auth/internal/domain/repositories"
	"github.com/video-platform/services/auth/internal/usecase/commands"
	"golang.org/x/crypto/bcrypt"
)

type registerUseCaseImpl struct {
	userRepo repositories.UserRepository
}

func NewRegisterUseCase(userRepo repositories.UserRepository) RegisterUseCase {
	return &registerUseCaseImpl{
		userRepo: userRepo,
	}
}

func (uc *registerUseCaseImpl) Execute(ctx context.Context, cmd commands.RegisterCommand) (*RegisterOutput, error) {
	existingUser, _ := uc.userRepo.FindByEmail(ctx, cmd.Email)
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	existingUser, _ = uc.userRepo.FindByUsername(ctx, cmd.Username)
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &entities.User{
		Username:     cmd.Username,
		Email:        cmd.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &RegisterOutput{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}
