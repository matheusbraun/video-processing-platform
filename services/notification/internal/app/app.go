package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"go.uber.org/fx"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/video-platform/services/notification/internal/controller"
	"github.com/video-platform/services/notification/internal/domain/repositories"
	"github.com/video-platform/services/notification/internal/infrastructure/messaging"
	"github.com/video-platform/services/notification/internal/infrastructure/persistence"
	"github.com/video-platform/services/notification/internal/infrastructure/smtp"
	"github.com/video-platform/services/notification/internal/usecase/sendemail"
	"github.com/video-platform/shared/pkg/config"
	"github.com/video-platform/shared/pkg/database/postgres"
	"github.com/video-platform/shared/pkg/messaging/rabbitmq"
	"github.com/video-platform/shared/pkg/metrics"
	"gorm.io/gorm"
)

func InitializeApp() *fx.App {
	return fx.New(
		fx.Provide(
			config.Load,

			func() *metrics.Metrics {
				return metrics.NewMetrics("notification")
			},

			func(cfg *config.Config) (*gorm.DB, error) {
				return postgres.NewPostgresDB(cfg.DatabaseURL)
			},

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
			go func() {
				log.Println("Starting metrics server on :8080")
				http.ListenAndServe(":8080", promhttp.Handler())
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
