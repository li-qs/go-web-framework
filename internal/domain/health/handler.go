package health

import (
	"context"
	"myframework/pkg/mysql"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

type Handler struct {
	db *mysql.DB
}

func NewHandler(db *mysql.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Liveness(c *echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func (h *Handler) Readiness(c *echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	if err := h.db.DB.PingContext(ctx); err != nil {
		return c.String(http.StatusServiceUnavailable, "unavailable")
	}
	return c.String(http.StatusOK, "OK")
}
