package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/anousonefs/golang-htmx-template/internal/config"
	"github.com/anousonefs/golang-htmx-template/internal/middleware"
	"github.com/anousonefs/golang-htmx-template/internal/user"
	"github.com/anousonefs/golang-htmx-template/internal/utils"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/discord"
	"github.com/markbates/goth/providers/github"

	"github.com/o1egl/paseto/v2"
	"github.com/sirupsen/logrus"
)

type service struct {
	user      user.Service
	pasetoKey []byte
}

func NewService(user user.Service, store sessions.Store, cfg config.Config) *service {

	gothic.Store = store

	goth.UseProviders(
		github.New(
			cfg.GetGithubClientID(),
			cfg.GetGithubClientID(),
			buildCallbackURL("github", cfg),
		),
		discord.New(
			cfg.GetDiscordClientID(),
			cfg.GetDiscordClientSecret(),
			buildCallbackURL("discord", cfg),
		),
	)
	return &service{user, cfg.GetPasetoSecret()}
}

func (s *service) StoreUserSession(c echo.Context, user goth.User) error {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := gothic.Store.Get(c.Request(), SessionName)

	session.Values["user"] = user

	err := session.Save(c.Request(), c.Response().Writer)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return nil
}

func (u service) Login(ctx context.Context, req LoginRequest) (res LoginResponse, err error) {
	user, err := u.user.GetUser(ctx, user.FilterUser{Username: req.Username})
	if err != nil {
		return LoginResponse{}, err
	}
	if err := utils.ComparePassword(req.Password, user.Password); err != nil {
		return LoginResponse{}, err
	}
	return generateToken(u.pasetoKey, user)
}

func (u service) RefreshToken(ctx context.Context, req RefreshTokenRequest) (res LoginResponse, err error) {
	defer func() {
		if err != nil {
			logrus.Errorf("u.RefreshToken: %v\n", err)
		}
	}()
	claims, err := u.verifyIDToken(ctx, req.RefreshToken)
	if err != nil {
		return LoginResponse{}, ErrUnProcessAbleEntity
	}
	var renewable bool
	if err := claims.Get("renewable", &renewable); err != nil || !renewable {
		return LoginResponse{}, ErrInternalServerError
	}
	user, err := u.user.GetUser(ctx, user.FilterUser{Username: claims.Subject})
	if err != nil {
		return LoginResponse{}, err
	}
	res, err = generateToken(u.pasetoKey, user)
	if err != nil {
		return LoginResponse{}, ErrInternalServerError
	}
	return res, nil
}

func (u *service) verifyIDToken(_ context.Context, idToken string) (paseto.JSONToken, error) {
	claims := paseto.JSONToken{}
	if err := paseto.Decrypt(idToken, u.pasetoKey, &claims, nil); err != nil {
		return claims, err
	}
	if err := claims.Validate(); err != nil {
		return claims, err
	}
	return claims, nil
}

var now = time.Now

func generateToken(secret []byte, u *user.UserDetail) (LoginResponse, error) {
	issAt := now()
	claims := paseto.JSONToken{
		Subject:    u.Username,
		IssuedAt:   issAt,
		Expiration: issAt.Add(5 * time.Hour),
		NotBefore:  issAt,
	}
	userClaims := middleware.UserClaim{
		ID:           u.ID,
		DepartmentID: u.DepartmentID,
		RoleID:       u.Role.ID,
	}
	claims.Set("user", userClaims)
	accessKey, err := paseto.Encrypt(secret, claims, nil)
	if err != nil {
		return LoginResponse{}, err
	}
	claims.Set("renewable", true)
	claims.Expiration = claims.Expiration.Add(48 * time.Hour)
	refreshKey, err := paseto.Encrypt(secret, claims, nil)
	if err != nil {
		return LoginResponse{}, err
	}
	return LoginResponse{
		AccessToken:  accessKey,
		RefreshToken: refreshKey,
	}, nil
}

func buildCallbackURL(provider string, cfg config.Config) string {
	return fmt.Sprintf("%s:%s/auth/callback?provider=%s", cfg.GetBaseUrl(), cfg.GetAppPort(), provider)
}
