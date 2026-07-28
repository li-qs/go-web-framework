package handler

import (
	"net/http"
	"net/url"

	"myapi/internal/dto"
	"myapi/internal/response"
	"myapi/internal/service"

	"github.com/labstack/echo/v5"
)

type Login struct {
	LoginService *service.Login
}

func (h *Login) Login(c *echo.Context) error {
	var in dto.Login
	if err := c.Bind(&in); err != nil {
		return response.JsonError(c, 400, "用户名或密码错误")
	}
	if err := c.Validate(&in); err != nil {
		return response.JsonError(c, 400, "用户名或密码错误")
	}
	user, isValid, err := h.LoginService.AuthUser(in.Username, in.Password)
	if err != nil || !isValid {
		return response.JsonError(c, 400, "用户名或密码错误")
	}

	token, maxAge, err := h.LoginService.GenerateToken(user)
	if err != nil {
		return response.JsonError(c, 500, "服务器错误")
	}
	c.SetCookie(&http.Cookie{
		Name:     "token",
		Value:    url.QueryEscape(token),
		MaxAge:   maxAge,
		Path:     "/",
		Domain:   "",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	return response.JsonData(c, "")
}

func (h *Login) Logout(c *echo.Context) error {
	c.SetCookie(&http.Cookie{
		Name:     "token",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		Domain:   "",
		SameSite: http.SameSiteStrictMode,
	})
	return response.JsonData(c, "")
}
