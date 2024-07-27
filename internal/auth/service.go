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
	"github.com/markbates/goth/providers/facebook"

	"github.com/o1egl/paseto/v2"
	"github.com/sirupsen/logrus"
)

type Service struct {
	user user.Service
	cfg  config.Config
}

func NewService(user user.Service, store sessions.Store, cfg config.Config) *Service {

	gothic.Store = store

	goth.UseProviders(
		facebook.New(
			cfg.FacebookAppID(),
			cfg.FacebookAppSecret(),
			buildCallbackURL("facebook", cfg),
		),
		discord.New(
			cfg.DiscordClientID(),
			cfg.DiscordClientSecret(),
			buildCallbackURL("discord", cfg),
		),
	)
	return &Service{user, cfg}
}

func (s Service) GetSessionUser(c echo.Context) (goth.User, error) {
	session, err := gothic.Store.Get(c.Request(), s.cfg.SessionName())
	if err != nil {
		return goth.User{}, err
	}

	u := session.Values["user"]
	if u == nil {
		return goth.User{}, fmt.Errorf("user is not authenticated! %v", u)
	}

	return u.(goth.User), nil
}

func (s *Service) SetCookie(c echo.Context, user goth.User) error {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, err := gothic.Store.Get(c.Request(), s.cfg.SessionName())
	if err != nil {
		logrus.Errorf("StoreUserSession.Get(): %v\n", err)
		return err
	}

	session.Values["user"] = user

	fmt.Printf("user: %#v\n", user)

	if err := session.Save(c.Request(), c.Response().Writer); err != nil {
		logrus.Errorf("StoreUserSession.Store(): %v\n", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	/* cookie := new(http.Cookie) */
	/* cookie.Name = s.cfg.SessionName() */
	/* cookie.Value = user.AccessToken */
	/* cookie.Expires = time.Now().Add(25 * time.Minute) */
	/* cookie.Path = "/" */
	/* cookie.HttpOnly = true */
	/**/
	/* c.SetCookie(cookie) */

	return nil
}

func (s Service) Login(ctx context.Context, req LoginRequest) (res LoginResponse, err error) {
	user, err := s.user.GetUser(ctx, user.FilterUser{Username: req.Username})
	if err != nil {
		return LoginResponse{}, err
	}
	if err := utils.ComparePassword(req.Password, user.Password); err != nil {
		return LoginResponse{}, err
	}
	return generateToken(s.cfg.PasetoSecret(), user)
}

func (s Service) RefreshToken(ctx context.Context, req RefreshTokenRequest) (res LoginResponse, err error) {
	defer func() {
		if err != nil {
			logrus.Errorf("u.RefreshToken: %v\n", err)
		}
	}()
	claims, err := s.verifyIDToken(ctx, req.RefreshToken)
	if err != nil {
		return LoginResponse{}, ErrUnProcessAbleEntity
	}
	var renewable bool
	if err := claims.Get("renewable", &renewable); err != nil || !renewable {
		return LoginResponse{}, ErrInternalServerError
	}
	user, err := s.user.GetUser(ctx, user.FilterUser{Username: claims.Subject})
	if err != nil {
		return LoginResponse{}, err
	}
	res, err = generateToken(s.cfg.PasetoSecret(), user)
	if err != nil {
		return LoginResponse{}, ErrInternalServerError
	}
	return res, nil
}

func (s *Service) verifyIDToken(_ context.Context, idToken string) (paseto.JSONToken, error) {
	claims := paseto.JSONToken{}
	if err := paseto.Decrypt(idToken, s.cfg.PasetoSecret(), &claims, nil); err != nil {
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
	return fmt.Sprintf("%s:%s/auth/callback?provider=%s", cfg.BaseUrl(), cfg.AppPort(), provider)
}
