package sendemail

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/video-platform/services/notification/internal/domain/entities"
	"github.com/video-platform/services/notification/internal/usecase/commands"
)

type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) Create(ctx context.Context, notification *entities.Notification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockNotificationRepository) MarkAsSent(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockNotificationRepository) MarkAsFailed(ctx context.Context, id int64, errorMessage string) error {
	args := m.Called(ctx, id, errorMessage)
	return args.Error(0)
}

type MockSMTPClient struct {
	mock.Mock
}

func (m *MockSMTPClient) SendEmail(to, subject, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

func TestSendEmailUseCase_Execute_CompletedSuccess(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockNotificationRepository)
	mockSMTP := new(MockSMTPClient)

	cmd := commands.SendEmailCommand{
		UserID:     1,
		UserEmail:  "test@example.com",
		VideoID:    "video-123",
		Status:     "COMPLETED",
		FrameCount: 100,
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.Notification")).Return(nil).Run(func(args mock.Arguments) {
		notification := args.Get(1).(*entities.Notification)
		notification.ID = 1
	})
	mockSMTP.On("SendEmail", "test@example.com", "Video Processing Completed", mock.Anything).Return(nil)
	mockRepo.On("MarkAsSent", ctx, int64(1)).Return(nil)

	useCase := NewSendEmailUseCase(mockRepo, mockSMTP)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockSMTP.AssertExpectations(t)
}

func TestSendEmailUseCase_Execute_FailedSuccess(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockNotificationRepository)
	mockSMTP := new(MockSMTPClient)

	cmd := commands.SendEmailCommand{
		UserID:       1,
		UserEmail:    "test@example.com",
		VideoID:      "video-123",
		Status:       "FAILED",
		ErrorMessage: "FFmpeg processing error",
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.Notification")).Return(nil).Run(func(args mock.Arguments) {
		notification := args.Get(1).(*entities.Notification)
		notification.ID = 1
	})
	mockSMTP.On("SendEmail", "test@example.com", "Video Processing Failed", mock.Anything).Return(nil)
	mockRepo.On("MarkAsSent", ctx, int64(1)).Return(nil)

	useCase := NewSendEmailUseCase(mockRepo, mockSMTP)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockSMTP.AssertExpectations(t)
}

func TestSendEmailUseCase_Execute_CreateNotificationError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockNotificationRepository)
	mockSMTP := new(MockSMTPClient)

	cmd := commands.SendEmailCommand{
		UserID:     1,
		UserEmail:  "test@example.com",
		VideoID:    "video-123",
		Status:     "COMPLETED",
		FrameCount: 100,
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.Notification")).Return(errors.New("database error"))

	useCase := NewSendEmailUseCase(mockRepo, mockSMTP)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create notification record")

	mockRepo.AssertExpectations(t)
}

func TestSendEmailUseCase_Execute_SendEmailError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockNotificationRepository)
	mockSMTP := new(MockSMTPClient)

	cmd := commands.SendEmailCommand{
		UserID:     1,
		UserEmail:  "test@example.com",
		VideoID:    "video-123",
		Status:     "COMPLETED",
		FrameCount: 100,
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.Notification")).Return(nil).Run(func(args mock.Arguments) {
		notification := args.Get(1).(*entities.Notification)
		notification.ID = 1
	})
	mockSMTP.On("SendEmail", "test@example.com", "Video Processing Completed", mock.Anything).Return(errors.New("SMTP error"))
	mockRepo.On("MarkAsFailed", ctx, int64(1), "SMTP error").Return(nil)

	useCase := NewSendEmailUseCase(mockRepo, mockSMTP)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send email")

	mockRepo.AssertExpectations(t)
	mockSMTP.AssertExpectations(t)
}

func TestSendEmailUseCase_Execute_MarkAsSentError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockNotificationRepository)
	mockSMTP := new(MockSMTPClient)

	cmd := commands.SendEmailCommand{
		UserID:     1,
		UserEmail:  "test@example.com",
		VideoID:    "video-123",
		Status:     "COMPLETED",
		FrameCount: 100,
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.Notification")).Return(nil).Run(func(args mock.Arguments) {
		notification := args.Get(1).(*entities.Notification)
		notification.ID = 1
	})
	mockSMTP.On("SendEmail", "test@example.com", "Video Processing Completed", mock.Anything).Return(nil)
	mockRepo.On("MarkAsSent", ctx, int64(1)).Return(errors.New("update error"))

	useCase := NewSendEmailUseCase(mockRepo, mockSMTP)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	// Email was sent successfully, so use case returns no error
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockSMTP.AssertExpectations(t)
}

func TestSendEmailUseCase_Execute_CompletedEmailContent(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockNotificationRepository)
	mockSMTP := new(MockSMTPClient)

	cmd := commands.SendEmailCommand{
		UserID:     1,
		UserEmail:  "test@example.com",
		VideoID:    "video-123",
		Status:     "COMPLETED",
		FrameCount: 100,
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.Notification")).Return(nil).Run(func(args mock.Arguments) {
		notification := args.Get(1).(*entities.Notification)
		notification.ID = 1
	})

	var capturedBody string
	mockSMTP.On("SendEmail", "test@example.com", "Video Processing Completed", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		capturedBody = args.String(2)
	})
	mockRepo.On("MarkAsSent", ctx, int64(1)).Return(nil)

	useCase := NewSendEmailUseCase(mockRepo, mockSMTP)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.Contains(t, capturedBody, "video-123")
	assert.Contains(t, capturedBody, "100")
	assert.Contains(t, capturedBody, "processed successfully")

	mockRepo.AssertExpectations(t)
	mockSMTP.AssertExpectations(t)
}

func TestSendEmailUseCase_Execute_FailedEmailContent(t *testing.T) {
	// Arrange
	ctx := context.Background()
	mockRepo := new(MockNotificationRepository)
	mockSMTP := new(MockSMTPClient)

	cmd := commands.SendEmailCommand{
		UserID:       1,
		UserEmail:    "test@example.com",
		VideoID:      "video-123",
		Status:       "FAILED",
		ErrorMessage: "FFmpeg processing error",
	}

	mockRepo.On("Create", ctx, mock.AnythingOfType("*entities.Notification")).Return(nil).Run(func(args mock.Arguments) {
		notification := args.Get(1).(*entities.Notification)
		notification.ID = 1
	})

	var capturedBody string
	mockSMTP.On("SendEmail", "test@example.com", "Video Processing Failed", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		capturedBody = args.String(2)
	})
	mockRepo.On("MarkAsSent", ctx, int64(1)).Return(nil)

	useCase := NewSendEmailUseCase(mockRepo, mockSMTP)

	// Act
	err := useCase.Execute(ctx, cmd)

	// Assert
	assert.NoError(t, err)
	assert.Contains(t, capturedBody, "video-123")
	assert.Contains(t, capturedBody, "FFmpeg processing error")
	assert.Contains(t, capturedBody, "processing failed")

	mockRepo.AssertExpectations(t)
	mockSMTP.AssertExpectations(t)
}
