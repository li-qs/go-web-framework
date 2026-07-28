package logger

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.uber.org/zap"
)

var Logger *zap.Logger

func Init() error {
	l, err := zap.NewProduction()
	if err != nil {
		return err
	}
	Logger = l
	return nil
}

func Sync() error {
	return Logger.Sync()
}

func Middleware() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			Logger.Info("request",
				zap.String("uri", v.URI),
				zap.Int("status", v.Status),
			)
			return nil
		},
	})
}
