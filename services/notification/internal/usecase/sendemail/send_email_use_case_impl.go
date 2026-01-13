package sendemail

import (
	"context"
	"fmt"

	"github.com/video-platform/services/notification/internal/domain/entities"
	"github.com/video-platform/services/notification/internal/domain/repositories"
	"github.com/video-platform/services/notification/internal/infrastructure/smtp"
	"github.com/video-platform/services/notification/internal/usecase/commands"
	"github.com/video-platform/shared/pkg/logging"
)

type sendEmailUseCaseImpl struct {
	notificationRepo repositories.NotificationRepository
	smtpClient       smtp.SMTPClient
}

func NewSendEmailUseCase(
	notificationRepo repositories.NotificationRepository,
	smtpClient smtp.SMTPClient,
) SendEmailUseCase {
	return &sendEmailUseCaseImpl{
		notificationRepo: notificationRepo,
		smtpClient:       smtpClient,
	}
}

func (uc *sendEmailUseCaseImpl) Execute(ctx context.Context, cmd commands.SendEmailCommand) error {
	var subject string
	var body string

	if cmd.Status == "COMPLETED" {
		subject = "Video Processing Completed"
		body = fmt.Sprintf(
			"Your video (ID: %s) has been processed successfully!\n\n"+
				"Frames extracted: %d\n"+
				"You can now download the ZIP file from the application.\n\n"+
				"Thank you for using our service!",
			cmd.VideoID, cmd.FrameCount,
		)
	} else {
		subject = "Video Processing Failed"
		body = fmt.Sprintf(
			"Unfortunately, your video (ID: %s) processing failed.\n\n"+
				"Error: %s\n\n"+
				"Please try uploading again or contact support.",
			cmd.VideoID, cmd.ErrorMessage,
		)
	}

	notification := &entities.Notification{
		UserID:    cmd.UserID,
		VideoID:   &cmd.VideoID,
		Type:      entities.TypeEmail,
		Status:    entities.StatusPending,
		Recipient: cmd.UserEmail,
		Subject:   &subject,
	}

	if err := uc.notificationRepo.Create(ctx, notification); err != nil {
		return fmt.Errorf("failed to create notification record: %w", err)
	}

	logging.Info("Sending email notification", "recipient", cmd.UserEmail, "video_id", cmd.VideoID)

	if err := uc.smtpClient.SendEmail(cmd.UserEmail, subject, body); err != nil {
		errMsg := err.Error()
		if updateErr := uc.notificationRepo.MarkAsFailed(ctx, notification.ID, errMsg); updateErr != nil {
			logging.Error("Failed to mark notification as failed", "error", updateErr)
		}
		return fmt.Errorf("failed to send email: %w", err)
	}

	if err := uc.notificationRepo.MarkAsSent(ctx, notification.ID); err != nil {
		logging.Error("Failed to mark notification as sent", "error", err)
	}

	logging.Info("Email sent successfully", "recipient", cmd.UserEmail, "video_id", cmd.VideoID)
	return nil
}
