package presenter

import (
	"github.com/video-platform/services/auth/internal/infrastructure/api/dto"
	"github.com/video-platform/services/auth/internal/usecase/login"
	"github.com/video-platform/services/auth/internal/usecase/refresh"
	"github.com/video-platform/services/auth/internal/usecase/register"
)

type authPresenterImpl struct{}

func NewAuthPresenter() AuthPresenter {
	return &authPresenterImpl{}
}

func (p *authPresenterImpl) PresentRegister(output *register.RegisterOutput) *dto.RegisterResponse {
	return &dto.RegisterResponse{
		UserID:   output.UserID,
		Username: output.Username,
		Email:    output.Email,
	}
}

func (p *authPresenterImpl) PresentLogin(output *login.LoginOutput) *dto.LoginResponse {
	return &dto.LoginResponse{
		AccessToken:  output.AccessToken,
		RefreshToken: output.RefreshToken,
		ExpiresIn:    output.ExpiresIn,
		User: dto.UserInfo{
			UserID:   output.UserID,
			Username: output.Username,
			Email:    output.Email,
		},
	}
}

func (p *authPresenterImpl) PresentRefresh(output *refresh.RefreshOutput) *dto.RefreshResponse {
	return &dto.RefreshResponse{
		AccessToken:  output.AccessToken,
		RefreshToken: output.RefreshToken,
		ExpiresIn:    output.ExpiresIn,
	}
}
