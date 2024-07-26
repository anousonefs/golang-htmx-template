package auth

import (
	"net/http"

	"github.com/anousonefs/golang-htmx-template/internal/config"

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

func (h handler) Install(app *echo.Echo) {
	v1 := app.Group("/api/v1")
	v1.POST("/login", h.login)
	v1.POST("/refresh-token", h.refreshToken)
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
