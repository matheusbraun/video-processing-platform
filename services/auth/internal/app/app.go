package app

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/video-platform/services/auth/internal/controller"
	"github.com/video-platform/services/auth/internal/domain/repositories"
	apiController "github.com/video-platform/services/auth/internal/infrastructure/api/controller"
	"github.com/video-platform/services/auth/internal/infrastructure/persistence"
	"github.com/video-platform/services/auth/internal/presenter"
	"github.com/video-platform/services/auth/internal/usecase/login"
	"github.com/video-platform/services/auth/internal/usecase/logout"
	"github.com/video-platform/services/auth/internal/usecase/refresh"
	"github.com/video-platform/services/auth/internal/usecase/register"
	"github.com/video-platform/shared/pkg/auth/jwt"
	"github.com/video-platform/shared/pkg/config"
	"github.com/video-platform/shared/pkg/database/postgres"
)

func InitializeApp() *fx.App {
	return fx.New(
		fx.Provide(
			config.Load,
			postgres.NewPostgresDB,

			func(cfg *config.Config) jwt.JWTManager {
				return jwt.NewJWTManager(cfg.JWTSecret, cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry)
			},

			fx.Annotate(persistence.NewUserRepository, fx.As(new(repositories.UserRepository))),
			fx.Annotate(persistence.NewRefreshTokenRepository, fx.As(new(repositories.RefreshTokenRepository))),

			fx.Annotate(register.NewRegisterUseCase, fx.As(new(register.RegisterUseCase))),
			fx.Annotate(login.NewLoginUseCase, fx.As(new(login.LoginUseCase))),
			fx.Annotate(refresh.NewRefreshUseCase, fx.As(new(refresh.RefreshUseCase))),
			fx.Annotate(logout.NewLogoutUseCase, fx.As(new(logout.LogoutUseCase))),

			fx.Annotate(controller.NewAuthController, fx.As(new(controller.AuthController))),
			fx.Annotate(presenter.NewAuthPresenter, fx.As(new(presenter.AuthPresenter))),

			chi.NewRouter,
			apiController.NewAuthHTTPController,
		),
		fx.Invoke(registerRoutes),
		fx.Invoke(startHTTPServer),
	)
}

func registerRoutes(r *chi.Mux, httpController *apiController.AuthHTTPController) {
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	r.Route("/api/v1/auth", func(r chi.Router) {
		httpController.RegisterRoutes(r)
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
				log.Printf("Starting Auth Service on %s", addr)
				if err := http.ListenAndServe(addr, r); err != nil {
					log.Fatalf("Failed to start HTTP server: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Shutting down Auth Service gracefully")
			return nil
		},
	})
}
