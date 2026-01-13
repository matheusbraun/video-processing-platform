package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/fx"

	"github.com/video-platform/services/processing-worker/internal/controller"
	"github.com/video-platform/services/processing-worker/internal/domain/repositories"
	"github.com/video-platform/services/processing-worker/internal/infrastructure/ffmpeg"
	"github.com/video-platform/services/processing-worker/internal/infrastructure/messaging"
	"github.com/video-platform/services/processing-worker/internal/infrastructure/persistence"
	"github.com/video-platform/services/processing-worker/internal/usecase/process"
	"github.com/video-platform/shared/pkg/config"
	"github.com/video-platform/shared/pkg/database/postgres"
	"github.com/video-platform/shared/pkg/messaging/rabbitmq"
	"github.com/video-platform/shared/pkg/storage/s3"
)

func InitializeApp() *fx.App {
	return fx.New(
		fx.Provide(
			config.Load,
			postgres.NewPostgresDB,

			func(cfg *config.Config) (s3.S3Client, error) {
				return s3.NewS3Client(cfg.AWSRegion, cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey, cfg.S3UploadsBucket)
			},

			func(cfg *config.Config) (*rabbitmq.Consumer, error) {
				return rabbitmq.NewConsumer(cfg.RabbitMQURL)
			},

			func(cfg *config.Config) (rabbitmq.Publisher, error) {
				return rabbitmq.NewPublisher(cfg.RabbitMQURL)
			},

			ffmpeg.NewFFmpegService,

			fx.Annotate(persistence.NewVideoRepository, fx.As(new(repositories.VideoRepository))),

			func(
				videoRepo repositories.VideoRepository,
				s3Client s3.S3Client,
				ffmpegService ffmpeg.FFmpegService,
				publisher rabbitmq.Publisher,
				cfg *config.Config,
			) process.ProcessUseCase {
				return process.NewProcessUseCase(videoRepo, s3Client, ffmpegService, publisher, cfg.S3ProcessedBucket)
			},

			fx.Annotate(controller.NewWorkerController, fx.As(new(controller.WorkerController))),

			messaging.NewVideoConsumer,
		),
		fx.Invoke(startWorker),
	)
}

func startWorker(lc fx.Lifecycle, consumer *messaging.VideoConsumer) {
	ctx, cancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				log.Println("Starting Processing Worker")
				if err := consumer.Start(ctx); err != nil {
					log.Printf("Worker error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(_ context.Context) error {
			log.Println("Shutting down Processing Worker")
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
