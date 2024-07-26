package auth

import (
	"fmt"
	"net/http"

	"github.com/anousonefs/golang-htmx-template/internal/auth/views"
	"github.com/anousonefs/golang-htmx-template/internal/config"
	"github.com/markbates/goth/gothic"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
)

type handler struct {
	auth *service
}

func NewHandler(app *echo.Echo, auth *service, cfg config.Config) *handler {
	return &handler{
		auth,
	}
}

func (h handler) Install(e *echo.Echo) {
	v1 := e.Group("/api/v1")
	v1.POST("/login", h.login)
	v1.POST("/refresh-token", h.refreshToken)

	e.GET("/auth/:provider", h.providerLogin)
	e.GET("/auth/:provider/callback", h.authCallback)

	e.GET("/login", h.loginPage)
}

func (h handler) loginPage(c echo.Context) error {
	if err := views.Login().Render(c.Request().Context(), c.Response().Writer); err != nil {
		return err
	}
	return nil
}

func (h handler) login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		logrus.Errorf("bind: %v\n", err)
		hs := HttpStatusPbFromRPC(StatusBindingFailure)
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(int(hs.Error.Code), b)
	}
	ctx := c.Request().Context()
	res, err := h.auth.Login(ctx, req)
	if err != nil {
		hs := HttpStatusPbFromRPC(GRPCStatusFromErr(err))
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(int(hs.Error.Code), b)
	}
	return c.JSON(http.StatusOK, res)
}

func (h handler) refreshToken(c echo.Context) error {
	var req RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		logrus.Errorf("bind: %v\n", err)
		hs := HttpStatusPbFromRPC(StatusBindingFailure)
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(int(hs.Error.Code), b)
	}
	ctx := c.Request().Context()
	res, err := h.auth.RefreshToken(ctx, req)
	if err != nil {
		hs := HttpStatusPbFromRPC(GRPCStatusFromErr(err))
		b, _ := protojson.Marshal(hs)
		return c.JSONBlob(int(hs.Error.Code), b)
	}
	return c.JSON(http.StatusOK, res)
}

func (h handler) providerLogin(c echo.Context) error {
	// try to get the user without re-authenticating
	if _, err := gothic.CompleteUserAuth(c.Response().Writer, c.Request()); err == nil {
		if err := views.Login().Render(c.Request().Context(), c.Response().Writer); err != nil {
			return err
		}
	} else {
		gothic.BeginAuthHandler(c.Response().Writer, c.Request())
	}
	return nil
}

func (h handler) authCallback(c echo.Context) error {
	// Use the Response and Request from Echo context
	_, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintln(err))
	}

	/* err = h.auth.StoreUserSession(c.Response(), c.Request(), user) */
	/* if err != nil { */
	/* 	log.Println(err) */
	/* 	return c.String(http.StatusInternalServerError, "Error storing user session") */
	/* } */

	return c.Redirect(http.StatusTemporaryRedirect, "/")
}
