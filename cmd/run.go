package cmd

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/anousonefs/golang-htmx-template/internal/activity"
	"github.com/anousonefs/golang-htmx-template/internal/auth"
	"github.com/anousonefs/golang-htmx-template/internal/config"
	"github.com/anousonefs/golang-htmx-template/internal/home"
	mdw "github.com/anousonefs/golang-htmx-template/internal/middleware"
	"github.com/anousonefs/golang-htmx-template/internal/user"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"

	casbinPgAdapter "github.com/cychiuae/casbin-pg-adapter"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

//go:embed policy.conf
var fs embed.FS

func Run() error {
	cfg, err := config.NewConfig()
	if err != nil {
		return err
	}
	ctx := context.Background()

	errCh := make(chan error, 1)
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer cancel()

	db, err := sql.Open(cfg.GetDBDriver(), cfg.DSNInfo())
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}
	defer db.Close()

	e := newEchoServer(cfg)

	adapter, err := casbinPgAdapter.NewAdapter(db, "permissions")
	if err != nil {
		return err
	}

	mc, err := fs.ReadFile("policy.conf")
	if err != nil {
		return err
	}
	model := string(mc)
	authz := mdw.New(mdw.Config{
		ModelFilePath: model,
		PolicyAdapter: adapter,
		Lookup: func(c echo.Context) string {
			return mdw.UserClaimFromContext(c.Request().Context()).RoleID
		},
		Forbidden: func(c echo.Context) error {
			return err
		},
	})

	activityRepo := activity.NewRepo(db)
	activityService := activity.NewService(activityRepo)

	repo := user.NewRepo(db, model, adapter, authz)
	userService := user.NewService(repo, activityService)
	user.NewHandler(e, userService, cfg).Install(e, cfg)

	sessionStore := auth.NewCookieStore(auth.SessionOptions{
		CookiesKey: "mycookies7898",
		MaxAge:     1000,
		Secure:     false,
		HttpOnly:   false,
	})

	authService := auth.NewService(userService, sessionStore, cfg)
	auth.NewHandler(e, authService, cfg).Install(e)

	homeService := home.NewService()
	home.NewHandler(e, homeService).Install(e)

	go func() {
		errCh <- e.Start(":" + cfg.GetAppPort())
	}()

	select {
	case <-ctx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := e.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown server failure: %v", err)
		}
	case err := <-errCh:
		return fmt.Errorf("start server error: %v", err)
	}
	return nil
}

func newEchoServer(_ config.Config) *echo.Echo {
	mws := []echo.MiddlewareFunc{
		middleware.LoggerWithConfig(middleware.LoggerConfig{
			Skipper: func(c echo.Context) bool {
				return c.Path() == "/" || c.Path() == "/_healthz"
			},
		}),
		middleware.Recover(),
		middleware.Secure(),
		middleware.CORS(),
	}
	e := echo.New()
	e.Use(mdw.CSPMiddleware)
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))

	pwd, _ := os.Getwd()
	e.Static("static", fmt.Sprintf("%v/static", pwd))

	e.HideBanner = true
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(mws...)
	e.GET("/_healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{"serverStatus": "running"})
	})

	return e
}
