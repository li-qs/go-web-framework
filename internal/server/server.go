package server

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

var defaultServerConfig = ServerConfig{
	Address:         ":8080",
	GracefulTimeout: 5 * time.Second,
}

type Server struct {
	cfg         ServerConfig
	echo        *echo.Echo
	routerGroup *echo.Group
}

type ServerConfig struct {
	Address         string
	CORSConfig      *middleware.CORSConfig
	GracefulTimeout time.Duration
}

type Option func(*ServerConfig)

func Address(addr string) Option {
	return func(c *ServerConfig) {
		c.Address = addr
	}
}

func WithCORS(config *middleware.CORSConfig) Option {
	return func(c *ServerConfig) {
		c.CORSConfig = config
	}
}

func WithGracefulTimeout(d time.Duration) Option {
	return func(c *ServerConfig) {
		c.GracefulTimeout = d
	}
}

func New(opts ...Option) *Server {
	cfg := ServerConfig{}

	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.Address == "" {
		cfg.Address = defaultServerConfig.Address
	}

	if cfg.GracefulTimeout == 0 {
		cfg.GracefulTimeout = defaultServerConfig.GracefulTimeout
	}

	e := echo.NewWithConfig(echo.Config{
		JSONSerializer: &JSONSerializer{},
		Validator:      &Validator{validator: validator.New()},
	})

	e.Use(middleware.Recover())
	e.Use(middleware.RequestLogger())
	e.Use(middleware.RequestID())

	if cfg.CORSConfig != nil {
		e.Use(middleware.CORSWithConfig(*cfg.CORSConfig))
	}

	return &Server{
		cfg:         cfg,
		echo:        e,
		routerGroup: e.Group(""),
	}
}

func (s *Server) Router() *echo.Group {
	return s.routerGroup
}

func (s *Server) Start(ctx context.Context) error {
	sc := echo.StartConfig{
		Address:         s.cfg.Address,
		GracefulTimeout: s.cfg.GracefulTimeout,
		HideBanner:      true,
	}
	return sc.Start(ctx, s.echo)
}
