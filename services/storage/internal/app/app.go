package app

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/video-platform/services/storage/internal/controller"
	"github.com/video-platform/services/storage/internal/infrastructure/api"
	"github.com/video-platform/services/storage/internal/usecase/createzip"
	"github.com/video-platform/shared/pkg/config"
	"github.com/video-platform/shared/pkg/storage/s3"
)

func InitializeApp() *fx.App {
	return fx.New(
		fx.Provide(
			config.Load,

			func(cfg *config.Config) (s3.S3Client, error) {
				return s3.NewS3Client(cfg.AWSRegion, cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey, cfg.S3ProcessedBucket)
			},

			fx.Annotate(createzip.NewCreateZipUseCase, fx.As(new(createzip.CreateZipUseCase))),
			fx.Annotate(controller.NewStorageController, fx.As(new(controller.StorageController))),

			chi.NewRouter,
			api.NewStorageHTTPController,
		),
		fx.Invoke(registerRoutes),
		fx.Invoke(startHTTPServer),
	)
}

func registerRoutes(r *chi.Mux, httpController *api.StorageHTTPController) {
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	httpController.RegisterRoutes(r)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

func startHTTPServer(lc fx.Lifecycle, r *chi.Mux, cfg *config.Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				addr := fmt.Sprintf(":%s", cfg.ServerPort)
				log.Printf("Starting Storage Service on %s", addr)
				if err := http.ListenAndServe(addr, r); err != nil {
					log.Fatalf("Failed to start HTTP server: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Shutting down Storage Service gracefully")
			return nil
		},
	})
}
