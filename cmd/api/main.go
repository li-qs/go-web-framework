package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"myframework/ent"
	"myframework/internal/config"
	"myframework/internal/domain/health"
	"myframework/internal/domain/user"
	myMiddleware "myframework/internal/middleware"
	"myframework/internal/server"

	_ "github.com/go-sql-driver/mysql"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var configFile string
	flag.StringVar(&configFile, "config", "./config.yaml", "config file path")
	flag.Parse()

	cfg, err := config.LoadFile(configFile)
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	slogOpts := &slog.HandlerOptions{}
	if os.Getenv("ENV") == "development" {
		slogOpts.Level = slog.LevelDebug
	} else {
		slogOpts.Level = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, slogOpts)))

	s := server.New(
		server.Address(cfg.ServerAddr),
		server.WithCORS(&middleware.CORSConfig{
			AllowOrigins:     cfg.AllowOrigins,
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
			AllowHeaders:     []string{"Content-Type", "Authorization", "Origin", "Accept", "X-Request-ID"},
			AllowCredentials: true,
			MaxAge:           86400,
		}),
	)

	client, err := ent.Open("mysql", cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("failed opening connection to mysql: %v", err)
	}
	defer client.Close()

	healthHandler := health.New(client)
	noAuth := s.Router()
	{
		noAuth.GET("/health", healthHandler.Liveness)
		noAuth.GET("/ready", healthHandler.Readiness)
	}

	userHandler := user.New(cfg, client)
	api := s.Router().Group("/api")
	{
		api.POST("/login", userHandler.Login, middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
			Store: middleware.NewRateLimiterMemoryStore(10),
			IdentifierExtractor: func(c *echo.Context) (string, error) {
				return c.RealIP(), nil
			},
		}))
		api.POST("/logout", userHandler.Logout)
		api.POST("/refresh", userHandler.RefreshToken)
	}

	authApi := s.Router().Group("/api")
	{
		authApi.Use(myMiddleware.Auth(cfg.JWTSecret))

		authApi.GET("/user", userHandler.UserInfo)
		authApi.PUT("/user/password", userHandler.UpdatePassword)
	}

	if err := s.Start(ctx); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
