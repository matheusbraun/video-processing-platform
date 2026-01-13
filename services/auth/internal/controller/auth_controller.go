package controller

import (
	"context"

	"github.com/video-platform/services/auth/internal/usecase/commands"
	"github.com/video-platform/services/auth/internal/usecase/login"
	"github.com/video-platform/services/auth/internal/usecase/refresh"
	"github.com/video-platform/services/auth/internal/usecase/register"
)

type AuthController interface {
	Register(ctx context.Context, cmd commands.RegisterCommand) (*register.RegisterOutput, error)
	Login(ctx context.Context, cmd commands.LoginCommand) (*login.LoginOutput, error)
	Refresh(ctx context.Context, cmd commands.RefreshCommand) (*refresh.RefreshOutput, error)
	Logout(ctx context.Context, cmd commands.LogoutCommand) error
}
