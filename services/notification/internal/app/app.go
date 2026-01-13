package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"go.uber.org/fx"

	"github.com/video-platform/services/notification/internal/controller"
	"github.com/video-platform/services/notification/internal/domain/repositories"
	"github.com/video-platform/services/notification/internal/infrastructure/messaging"
	"github.com/video-platform/services/notification/internal/infrastructure/persistence"
	"github.com/video-platform/services/notification/internal/infrastructure/smtp"
	"github.com/video-platform/services/notification/internal/usecase/sendemail"
	"github.com/video-platform/shared/pkg/config"
	"github.com/video-platform/shared/pkg/database/postgres"
	"github.com/video-platform/shared/pkg/messaging/rabbitmq"
)

func InitializeApp() *fx.App {
	return fx.New(
		fx.Provide(
			config.Load,
			postgres.NewPostgresDB,

			func(cfg *config.Config) (*rabbitmq.Consumer, error) {
				return rabbitmq.NewConsumer(cfg.RabbitMQURL)
			},

			func(cfg *config.Config) smtp.SMTPClient {
				return smtp.NewSMTPClient(
					cfg.SMTPHost,
					strconv.Itoa(cfg.SMTPPort),
					cfg.SMTPUser,
					cfg.SMTPPassword,
					cfg.SMTPUser,
				)
			},

			fx.Annotate(persistence.NewNotificationRepository, fx.As(new(repositories.NotificationRepository))),

			fx.Annotate(sendemail.NewSendEmailUseCase, fx.As(new(sendemail.SendEmailUseCase))),

			fx.Annotate(controller.NewNotificationController, fx.As(new(controller.NotificationController))),

			messaging.NewNotificationConsumer,
		),
		fx.Invoke(startWorker),
	)
}

func startWorker(lc fx.Lifecycle, consumer *messaging.NotificationConsumer) {
	ctx, cancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				log.Println("Starting Notification Service")
				if err := consumer.Start(ctx); err != nil {
					log.Printf("Worker error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(_ context.Context) error {
			log.Println("Shutting down Notification Service")
			cancel()
			return nil
		},
	})

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		cancel()
	}()
}
