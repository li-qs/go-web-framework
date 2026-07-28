package middleware

import (
	"net/url"

	"myapi/internal/response"
	"myapi/internal/service"

	"github.com/labstack/echo/v5"
)

func Auth(loginService *service.Login) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			tokenC, err := c.Cookie("token")
			if err != nil {
				return response.JsonError(c, 401, "请登录")
			}

			token, err := url.QueryUnescape(tokenC.Value)
			if err != nil {
				return response.JsonError(c, 401, "请登录")
			}

			u, ok, err := loginService.AuthToken(token)
			if err != nil || !ok {
				return response.JsonError(c, 401, "请登录")
			}

			u.PasswordHash = ""
			c.Set("user", u)

			return next(c)
		}
	}
}
