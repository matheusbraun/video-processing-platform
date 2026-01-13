package controller

import (
	"context"

	"github.com/video-platform/services/auth/internal/usecase/commands"
	"github.com/video-platform/services/auth/internal/usecase/login"
	"github.com/video-platform/services/auth/internal/usecase/logout"
	"github.com/video-platform/services/auth/internal/usecase/refresh"
	"github.com/video-platform/services/auth/internal/usecase/register"
)

type authControllerImpl struct {
	registerUseCase register.RegisterUseCase
	loginUseCase    login.LoginUseCase
	refreshUseCase  refresh.RefreshUseCase
	logoutUseCase   logout.LogoutUseCase
}

func NewAuthController(
	registerUseCase register.RegisterUseCase,
	loginUseCase login.LoginUseCase,
	refreshUseCase refresh.RefreshUseCase,
	logoutUseCase logout.LogoutUseCase,
) AuthController {
	return &authControllerImpl{
		registerUseCase: registerUseCase,
		loginUseCase:    loginUseCase,
		refreshUseCase:  refreshUseCase,
		logoutUseCase:   logoutUseCase,
	}
}

func (c *authControllerImpl) Register(ctx context.Context, cmd commands.RegisterCommand) (*register.RegisterOutput, error) {
	return c.registerUseCase.Execute(ctx, cmd)
}

func (c *authControllerImpl) Login(ctx context.Context, cmd commands.LoginCommand) (*login.LoginOutput, error) {
	return c.loginUseCase.Execute(ctx, cmd)
}

func (c *authControllerImpl) Refresh(ctx context.Context, cmd commands.RefreshCommand) (*refresh.RefreshOutput, error) {
	return c.refreshUseCase.Execute(ctx, cmd)
}

func (c *authControllerImpl) Logout(ctx context.Context, cmd commands.LogoutCommand) error {
	return c.logoutUseCase.Execute(ctx, cmd)
}
