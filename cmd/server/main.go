package main

import (
	"context"
	"myapi/internal/config"
	"myapi/internal/logger"
	"myapi/internal/router"
	"myapi/pkg/mysql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.uber.org/zap"
)

type Validator struct {
	validator *validator.Validate
}

func (v *Validator) Validate(i any) error {
	if err := v.validator.Struct(i); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}
	return nil
}

func main() {
	if err := logger.Init(); err != nil {
		panic(err)
	}
	defer logger.Logger.Sync()

	configFile := "./config.yaml"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	cfg, err := config.Load(configFile)
	if err != nil {
		logger.Logger.Fatal("load config failed", zap.Error(err))
	}

	db, err := mysql.InitDB(cfg.Database.Main)
	if err != nil {
		logger.Logger.Fatal("init mysql failed", zap.Error(err))
	}

	e := echo.New()
	e.Validator = &Validator{validator: validator.New()}

	e.Use(logger.Middleware())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.Server.AllowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "Origin", "Accept", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           86400,
	}))

	e.GET("/health", func(c *echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	router.Register(e, db, cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sc := echo.StartConfig{
		Address:         cfg.Server.ListenAddr,
		GracefulTimeout: 5 * time.Second,
	}
	if err := sc.Start(ctx, e); err != nil {
		logger.Logger.Fatal("failed to start server", zap.Error(err))
	}
}
