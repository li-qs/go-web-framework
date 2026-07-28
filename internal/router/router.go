package router

import (
	"myapi/internal/config"
	"myapi/internal/db"
	"myapi/internal/handler"
	myMiddleware "myapi/internal/middleware"
	"myapi/internal/repository"
	"myapi/internal/service"

	"github.com/labstack/echo/v5"
)

func Register(e *echo.Echo, cfg *config.Config) {
	userRepo := &repository.User{DB: db.DB}

	loginService := &service.Login{
		UserRepo:           userRepo,
		LoginExpireSeconds: cfg.Auth.LoginExpireSeconds,
	}

	userService := &service.User{
		UserRepo: userRepo,
	}

	login := handler.Login{LoginService: loginService}
	e.POST("/login", login.Login)
	e.POST("/logout", login.Logout)

	protected := e.Group("")
	protected.Use(myMiddleware.Auth(loginService))

	user := handler.User{UserService: userService, LoginService: loginService}
	protected.GET("/user", user.Get)
	protected.PUT("/user/password", user.UpdatePassword)
}
