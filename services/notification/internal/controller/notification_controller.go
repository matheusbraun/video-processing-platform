package controller

import (
	"context"

	"github.com/video-platform/services/notification/internal/usecase/commands"
	"github.com/video-platform/services/notification/internal/usecase/sendemail"
)

type NotificationController interface {
	SendEmail(ctx context.Context, cmd commands.SendEmailCommand) error
}

type notificationControllerImpl struct {
	sendEmailUseCase sendemail.SendEmailUseCase
}

func NewNotificationController(sendEmailUseCase sendemail.SendEmailUseCase) NotificationController {
	return &notificationControllerImpl{
		sendEmailUseCase: sendEmailUseCase,
	}
}

func (c *notificationControllerImpl) SendEmail(ctx context.Context, cmd commands.SendEmailCommand) error {
	return c.sendEmailUseCase.Execute(ctx, cmd)
}
