package home

import (
	"github.com/anousonefs/golang-htmx-template/internal/config"
	"github.com/anousonefs/golang-htmx-template/internal/dashboard/views"
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
	e.GET("/dashboard", h.dashboardPage, middleware.ValidateCookie(cfg)...)
}

func (h *handler) homePage(c echo.Context) error {
	comp := views.DashboardPage()
	if err := templates.Layout(comp, "hello").Render(c.Request().Context(), c.Response().Writer); err != nil {
		return err
	}
	return nil
}

func (h *handler) dashboardPage(c echo.Context) error {
	if err := views.DashboardPage().Render(c.Request().Context(), c.Response().Writer); err != nil {
		return err
	}
	return nil
}
