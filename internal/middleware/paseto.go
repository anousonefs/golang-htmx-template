package middleware

import (
	"fmt"

	"log"
	"net/http"
	"strings"
	"time"

	"github.com/anousonefs/golang-htmx-template/internal/config"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/sirupsen/logrus"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/o1egl/paseto/v2"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	// The size of the key used by ChaCha20-Poly1305 AEAD for PASETO's encryption and decryption, in bytes.
	keySize = 32
)

var (
	ErrPASETOUnsupported = echo.NewHTTPError(http.StatusBadRequest, "unsupported paseto version/purpose")
	ErrPASETOMissing     = echo.NewHTTPError(http.StatusBadRequest, "missing or malformed paseto")
)

var DefaultPASETOConfig = PASETOConfig{
	Skipper:     middleware.DefaultSkipper,
	ContextKey:  "user",
	TokenLookUp: "header:" + echo.HeaderAuthorization,
	AuthScheme:  "Bearer",
	Validators:  []paseto.Validator{},
}

type (
	PASETOSuccessHandler          func(echo.Context)
	PASETOErrorHandlerWithContext func(error, echo.Context) error
	pasetoExtractor               func(echo.Context) (string, error)
)

type PASETOConfig struct {
	Skipper                 middleware.Skipper
	SuccessHandler          PASETOSuccessHandler
	ErrorHandlerWithContext PASETOErrorHandlerWithContext

	// Signing key to validate token.
	SigningKey []byte

	Validators []paseto.Validator

	// The key to store user information from the token into context.
	// Optional. Default value is "user".
	ContextKey string

	// TokenLookUp
	//
	// A string in the form of "<source>:<name>" that is used
	// to extract token from the request.
	//
	// Optional. Default value is "header:Authorization".
	//
	// Possible values:
	//	- "header:<name>"
	// 	- "query:<name>"
	//	- "param:<name>"
	// 	- "cookie:<name>"
	TokenLookUp string

	// AuthScheme
	//
	// The scheme to be used in the Authorization header.
	//
	// Optional. Default value is "Bearer".
	AuthScheme string

	// The Redis to use for checking blacklist token.
	/* Redis *redis.Client */
}

func CheckCookie(sesssionName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if _, err := GetSessionUser(c.Request(), sesssionName); err != nil {
				logrus.Errorf("GetSessionUser(): %v\n", err)
				return c.Redirect(http.StatusTemporaryRedirect, "/login")
			}
			return next(c)
		}
	}
}

func PASETOWithConfig(config PASETOConfig) echo.MiddlewareFunc {
	if len(config.SigningKey) != keySize {
		log.Fatal("SigningKey must be 32 bytes length")
	}
	/* if config.Redis == nil { */
	/* 	log.Fatal("redis client must not be nil") */
	/* } */
	if config.Skipper == nil {
		config.Skipper = DefaultPASETOConfig.Skipper
	}
	if config.ContextKey == "" {
		config.ContextKey = DefaultPASETOConfig.ContextKey
	}
	if config.Validators == nil {
		config.Validators = DefaultPASETOConfig.Validators
	}
	if config.AuthScheme == "" {
		config.AuthScheme = DefaultPASETOConfig.AuthScheme
	}
	if config.TokenLookUp == "" {
		config.TokenLookUp = DefaultPASETOConfig.TokenLookUp
	}

	parts := strings.Split(config.TokenLookUp, ":")
	extractor := pasetoFromHeader(parts[1], config.AuthScheme)
	switch parts[0] {
	case "query":
		extractor = pasetoFromQuery(parts[1])
	case "param":
		extractor = pasetoFromParam(parts[1])
	case "cookie":
		extractor = pasetoFromCookie(parts[1])
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {

			if config.Skipper(c) {
				return next(c)
			}
			defer func() {
				if err != nil && config.ErrorHandlerWithContext != nil {
					err = config.ErrorHandlerWithContext(err, c)
				}
			}()

			auth, err := extractor(c)
			if err != nil {
				fmt.Printf("extractor error: %v\n", err)
				return err
			}
			if !strings.HasPrefix(auth, "v2.local") {
				fmt.Printf("not support token: %v\n", err)
				return ErrPASETOUnsupported
			}

			/* isBlackList, err := config.Redis.Exists(c.Request().Context(), auth).Result() */
			/* if err == nil && isBlackList == 1 { */
			/* 	return &echo.HTTPError{ */
			/* 		Code:    http.StatusUnauthorized, */
			/* 		Message: "invalid or expired token", */
			/* 	} */
			/* } */

			var claims paseto.JSONToken
			err = paseto.Decrypt(auth, config.SigningKey, &claims, nil)
			if err == nil {
				err = claims.Validate(append(config.Validators, paseto.ValidAt(time.Now()))...)
				if err == nil {
					c.Set(config.ContextKey, claims)
					if config.SuccessHandler != nil {
						config.SuccessHandler(c)
					}
					return next(c)
				}
			}
			println("invalid or expired token")
			return ErrUnauthorized
		}
	}
}

func PasetoFromHeader(c echo.Context) (string, error) {
	auth := c.Request().Header.Get(echo.HeaderAuthorization)
	authScheme := DefaultPASETOConfig.AuthScheme
	l := len(authScheme)
	if len(auth) > l+1 && auth[:l] == authScheme {
		return auth[l+1:], nil
	}
	return "", ErrPASETOMissing
}

func pasetoFromHeader(header string, authScheme string) pasetoExtractor {
	return func(c echo.Context) (string, error) {
		auth := c.Request().Header.Get(header)
		l := len(authScheme)
		if len(auth) > l+1 && auth[:l] == authScheme {
			return auth[l+1:], nil
		}
		return "", ErrPASETOMissing
	}
}

func pasetoFromQuery(param string) pasetoExtractor {
	return func(c echo.Context) (string, error) {
		token := c.QueryParam(param)
		if token == "" {
			return "", ErrPASETOMissing
		}
		return token, nil
	}
}

func pasetoFromParam(param string) pasetoExtractor {
	return func(c echo.Context) (string, error) {
		token := c.Param(param)
		if token == "" {
			return "", ErrPASETOMissing
		}
		return token, nil
	}
}

func pasetoFromCookie(name string) pasetoExtractor {
	return func(c echo.Context) (string, error) {
		cookie, err := c.Cookie(name)
		if err != nil {
			return "", ErrPASETOMissing
		}
		return cookie.Value, nil
	}
}

func Auth(cfg config.Config) []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
		PASETOWithConfig(PASETOConfig{
			Skipper:    func(c echo.Context) bool { return c.Path() == "/_healthz" },
			SigningKey: cfg.PasetoSecret(),
			ErrorHandlerWithContext: func(err error, c echo.Context) error {
				httpStatus := HttpStatusPbFromRPC(GRPCStatusFromErr(err))
				b, _ := protojson.Marshal(httpStatus)
				return c.JSONBlob(int(httpStatus.Error.Code), b)
			},
		},
		),
		SetClaimsMiddleware(),
	}
}

func ValidateCookie(cfg config.Config) []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
		CheckCookie(cfg.SessionName()),
	}
}

func GetSessionUser(r *http.Request, sessionName string) (goth.User, error) {
	session, err := gothic.Store.Get(r, sessionName)
	if err != nil {
		return goth.User{}, err
	}

	u := session.Values["user"]
	if u == nil {
		return goth.User{}, fmt.Errorf("user is not authenticated! %v", u)
	}

	return u.(goth.User), nil
}
