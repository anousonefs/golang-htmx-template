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
	auth *Service
}

func NewHandler(app *echo.Echo, auth *Service, cfg config.Config) *handler {
	return &handler{
		auth,
	}
}

func (h handler) Install(e *echo.Echo) {
	v1 := e.Group("/api/v1")
	v1.POST("/login", h.login)
	v1.POST("/refresh-token", h.refreshToken)

	e.GET("/auth", h.providerLogin)
	e.GET("/auth/callback", h.authCallback)

	e.GET("/login", h.loginPage)
	e.POST("/web/login", h.loginWeb)
}

func (h handler) loginPage(c echo.Context) error {
	if err := views.Login().Render(c.Request().Context(), c.Response().Writer); err != nil {
		return err
	}
	return nil
}

func (h handler) loginWeb(c echo.Context) error {
	req := LoginRequest{
		Email:    c.FormValue("email"),
		Password: c.FormValue("password"),
	}
	if req.Email == "" || req.Password == "" {
		hs := HttpStatusPbFromRPC(StatusBadRequest)
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

	if err := h.auth.SetCookie(c, res); err != nil {
		logrus.Printf("loginWeb.StoreUserSession(): %v\n", err)
		return c.String(http.StatusInternalServerError, "Error storing user session")
	}

	return c.NoContent(200)
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
	user, err := gothic.CompleteUserAuth(c.Response().Writer, c.Request())
	fmt.Printf("user: %#v\n", user)
	if err == nil {
		if err := views.Login().Render(c.Request().Context(), c.Response().Writer); err != nil {
			logrus.Errorf("call view(): %v\n", err)
			return err
		}
	} else {
		logrus.Errorf("CompleteUserAuth error: %v\n", err)
		gothic.BeginAuthHandler(c.Response().Writer, c.Request())
	}
	return nil
}

func (h handler) authCallback(c echo.Context) error {
	user, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		logrus.Errorf("authCallback.CompleteUserAuth(): %v\n", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}
	tokens, err := h.auth.genToken(c.Request().Context(), user.Email)
	if err != nil {
		return err
	}

	if err := h.auth.SetCookie(c, tokens); err != nil {
		logrus.Printf("authCallback.StoreUserSession(): %v\n", err)
		return c.String(http.StatusInternalServerError, "Error storing user session")
	}

	return c.Redirect(http.StatusTemporaryRedirect, "/")
}
