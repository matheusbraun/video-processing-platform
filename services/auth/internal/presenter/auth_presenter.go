package presenter

import (
	"github.com/video-platform/services/auth/internal/infrastructure/api/dto"
	"github.com/video-platform/services/auth/internal/usecase/login"
	"github.com/video-platform/services/auth/internal/usecase/refresh"
	"github.com/video-platform/services/auth/internal/usecase/register"
)

type AuthPresenter interface {
	PresentRegister(output *register.RegisterOutput) *dto.RegisterResponse
	PresentLogin(output *login.LoginOutput) *dto.LoginResponse
	PresentRefresh(output *refresh.RefreshOutput) *dto.RefreshResponse
}
