package home

import (
	"github.com/anousonefs/golang-htmx-template/internal/config"
	"github.com/anousonefs/golang-htmx-template/internal/home/views"
	"github.com/anousonefs/golang-htmx-template/internal/middleware"
	"github.com/anousonefs/golang-htmx-template/internal/templates"

	"github.com/labstack/echo/v4"
)

type handler struct {
	home Service
}

func NewHandler(e *echo.Echo, home Service) *handler {
	return &handler{
		home,
	}
}

func (h *handler) Install(e *echo.Echo, cfg config.Config) {
	e.GET("/", h.homePage, middleware.ValidateCookie(cfg)...)
}

func (h *handler) homePage(c echo.Context) error {
	comp := views.HomePage()
	if err := templates.Layout(comp, "My website").Render(c.Request().Context(), c.Response().Writer); err != nil {
		return err
	}
	return nil
}
