package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"myframework/internal/config"
	"myframework/internal/domain/health"
	"myframework/internal/domain/user"
	"myframework/internal/infra/repo"
	myMiddleware "myframework/internal/middleware"
	"myframework/internal/server"
	"myframework/pkg/mysql"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", "./config.yaml", "config file path")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.LoadFile(configFile)
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	db, err := mysql.Connect(cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("connect mysql failed: %v", err)
	}
	defer db.Close()

	userRepo := repo.NewUserRepo(db)
	tokenRepo := repo.NewTokenRepo(db, int32(cfg.RefreshTTL))

	userService := user.NewService(userRepo, tokenRepo, &user.ServiceOptions{
		JWTSecret:                 cfg.JWTSecret,
		TokenSalt:                 cfg.TokenSalt,
		AccessTokenExpireSeconds:  cfg.AccessTTL,
		RefreshTokenExpireSeconds: cfg.RefreshTTL,
	})

	register := func(e *echo.Echo) {
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     cfg.AllowOrigins,
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
			AllowHeaders:     []string{"Content-Type", "Authorization", "Origin", "Accept", "X-Request-ID"},
			AllowCredentials: true,
			MaxAge:           86400,
		}))

		healthHandler := health.NewHandler(db)
		userHandler := user.NewHandler(userService, *cfg.CookieSecure)

		noAuth := e.Group("")
		noAuth.GET("/health", healthHandler.Liveness)
		noAuth.GET("/ready", healthHandler.Readiness)

		api := e.Group("/api")
		api.POST("/login", userHandler.Login, middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
			Store: middleware.NewRateLimiterMemoryStore(10),
			IdentifierExtractor: func(c *echo.Context) (string, error) {
				return c.RealIP(), nil
			},
		}))
		api.POST("/logout", userHandler.Logout)
		api.POST("/refresh", userHandler.RefreshToken)

		authApi := e.Group("/api")
		authApi.Use(myMiddleware.Auth(cfg.JWTSecret))

		authApi.GET("/user", userHandler.UserInfo)
		authApi.PUT("/user/password", userHandler.UpdatePassword)
	}

	if err := server.Run(ctx, cfg.ServerAddr, cfg.LogLevel, register); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
