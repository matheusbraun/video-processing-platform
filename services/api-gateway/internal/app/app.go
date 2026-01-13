package app

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/video-platform/services/api-gateway/internal/controller"
	"github.com/video-platform/services/api-gateway/internal/domain/repositories"
	apiController "github.com/video-platform/services/api-gateway/internal/infrastructure/api/controller"
	"github.com/video-platform/services/api-gateway/internal/infrastructure/persistence"
	"github.com/video-platform/services/api-gateway/internal/presenter"
	"github.com/video-platform/services/api-gateway/internal/usecase/download"
	"github.com/video-platform/services/api-gateway/internal/usecase/list"
	"github.com/video-platform/services/api-gateway/internal/usecase/status"
	"github.com/video-platform/services/api-gateway/internal/usecase/upload"
	"github.com/video-platform/shared/pkg/auth/jwt"
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

			func(cfg *config.Config) jwt.JWTManager {
				return jwt.NewJWTManager(cfg.JWTSecret, cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry)
			},

			func(cfg *config.Config) (s3.S3Client, error) {
				return s3.NewS3Client(cfg.AWSRegion, cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey, cfg.S3UploadsBucket)
			},

			func(cfg *config.Config) (rabbitmq.Publisher, error) {
				return rabbitmq.NewPublisher(cfg.RabbitMQURL)
			},

			fx.Annotate(persistence.NewVideoRepository, fx.As(new(repositories.VideoRepository))),

			fx.Annotate(upload.NewUploadUseCase, fx.As(new(upload.UploadUseCase))),
			fx.Annotate(list.NewListUseCase, fx.As(new(list.ListUseCase))),
			fx.Annotate(status.NewStatusUseCase, fx.As(new(status.StatusUseCase))),
			fx.Annotate(download.NewDownloadUseCase, fx.As(new(download.DownloadUseCase))),

			fx.Annotate(controller.NewVideoController, fx.As(new(controller.VideoController))),
			fx.Annotate(presenter.NewVideoPresenter, fx.As(new(presenter.VideoPresenter))),

			chi.NewRouter,
			apiController.NewVideoHTTPController,
		),
		fx.Invoke(registerRoutes),
		fx.Invoke(startHTTPServer),
	)
}

func registerRoutes(r *chi.Mux, httpController *apiController.VideoHTTPController, jwtManager jwt.JWTManager) {
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	r.Route("/api/v1", func(r chi.Router) {
		httpController.RegisterRoutes(r, jwtManager)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func startHTTPServer(lc fx.Lifecycle, r *chi.Mux, cfg *config.Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				addr := fmt.Sprintf(":%s", cfg.ServerPort)
				log.Printf("Starting API Gateway on %s", addr)
				if err := http.ListenAndServe(addr, r); err != nil {
					log.Fatalf("Failed to start HTTP server: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Shutting down API Gateway gracefully")
			return nil
		},
	})
}
