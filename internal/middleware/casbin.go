package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/anousonefs/golang-htmx-template/internal/utils"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	casbinpgadapter "github.com/cychiuae/casbin-pg-adapter"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type Config struct {
	ModelFilePath string
	PolicyAdapter *casbinpgadapter.Adapter
	Enforcer      *casbin.Enforcer
	Lookup        func(echo.Context) string
	Unauthorized  echo.HandlerFunc
	Forbidden     echo.HandlerFunc
}

type CasbinMiddleware struct {
	config Config
}

func New(config ...Config) *CasbinMiddleware {
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.Enforcer == nil {
		if cfg.ModelFilePath == "" {
			cfg.ModelFilePath = "./policy.conf"
		}
		m, _ := model.NewModelFromString(cfg.ModelFilePath)

		enforcer, err := casbin.NewEnforcer(m, cfg.PolicyAdapter)
		if err != nil {
			log.Fatalf("echo: Casbin middleware error -> %v", err)
		}

		cfg.Enforcer = enforcer
	}

	if cfg.Lookup == nil {
		cfg.Lookup = func(c echo.Context) string { return "" }
	}

	if cfg.Unauthorized == nil {
		cfg.Unauthorized = func(c echo.Context) error {
			return c.JSON(http.StatusUnauthorized, echo.Map{})
		}
	}

	if cfg.Forbidden == nil {
		cfg.Forbidden = func(c echo.Context) error {
			return c.JSON(http.StatusForbidden, echo.Map{})
		}
	}

	return &CasbinMiddleware{
		config: cfg,
	}
}

type validationRule int

const (
	matchAll validationRule = iota
	atLeastOne
)

var MatchAll = func(o *Options) {
	o.ValidationRule = matchAll
}

var AtLeastOne = func(o *Options) {
	o.ValidationRule = atLeastOne
}

type PermissionParserFunc func(str string) []string

func permissionParserWithSeperator(sep string) PermissionParserFunc {
	return func(str string) []string {
		return strings.Split(str, sep)
	}
}

func PermissionParserWithSeperator(sep string) func(o *Options) {
	return func(o *Options) {
		o.PermissionParser = permissionParserWithSeperator(sep)
	}
}

type Options struct {
	ValidationRule   validationRule
	PermissionParser PermissionParserFunc
}

func (cm *CasbinMiddleware) RequiresPermissions(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		url := c.Request().URL.Path
		fmt.Printf("=> url: %+v\n", url)
		switch {
		case c.Param("name") != "":
			url = replaceParam(url, c.Param("name"))
		case c.Param("id") != "":
			url = replaceParam(url, c.Param("id"))
		case c.Param("number") != "":
			url = replaceParam(url, c.Param("number"))
		case c.Param("username") != "":
			url = replaceParam(url, c.Param("username"))
		case c.Param("flightID") != "":
			url = replaceParam(url, c.Param("flightID"))
		}
		resource := url2resource[url]
		action := url2cmethod[url]
		httpmethod := c.Request().Method
		if action == "" {
			action = httpmethod2string[httpmethod]
		}

		sub := cm.config.Lookup(c)
		if len(sub) == 0 {
			return cm.config.Unauthorized(c)
		}
		vals := append([]string{sub}, resource, action)
		fmt.Printf("sub: %v\n", sub)
		fmt.Printf("=> vals: %v\n", vals)
		if ok, err := cm.config.Enforcer.Enforce(utils.StringSliceToInterfaceSlice(vals)...); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{})
		} else if !ok {
			println("start")
			// todo: create another endpoint for branch sorting
			fmt.Printf("u: %v, query: %v\n", url, c.QueryParam(("isSorting")))
			if url == "/api/v1/branches" && c.QueryParam("isSorting") == "true" {
				println("resource branch sorting")
				vals3 := append([]string{sub}, "branchSorting", "list")
				if ok, err := cm.config.Enforcer.Enforce(utils.StringSliceToInterfaceSlice(vals3)...); err != nil {
					return c.JSON(http.StatusInternalServerError, echo.Map{})
				} else if !ok {
					return cm.config.Forbidden(c)
				}
				return next(c)
			}
			if (resource == "vendor" && action == "list") || (resource == "branch" && action == "list") || (resource == "boxType" && action == "list") || (resource == "boxSize" && action == "list") {
				println("=> check permission register box!")
				vals2 := append([]string{sub}, "registerBox", "create")
				if ok, err := cm.config.Enforcer.Enforce(utils.StringSliceToInterfaceSlice(vals2)...); err != nil {
					return c.JSON(http.StatusInternalServerError, echo.Map{})
				} else if !ok {
					return cm.config.Forbidden(c)
				}
				return next(c)
			}
			println("=> policy not match\n")
			return cm.config.Forbidden(c)
		}
		return next(c)
	}
}

// RoutePermission tries to find the current subject and determine if the
// subject has the required permissions according to predefined Casbin policies.
// This method uses http Path and Method as object and action.
func (cm *CasbinMiddleware) RoutePermission(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// sub := cm.config.Lookup(c)
		sub := "userid"
		if len(sub) == 0 {
			return cm.config.Unauthorized(c)
		}

		if ok, err := cm.config.Enforcer.Enforce(sub, c.Request().URL.Path, c.Request().Method); err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{})
		} else if !ok {
			return cm.config.Forbidden(c)
		}

		return next(c)
	}
}

// RequiresRoles tries to find the current subject and determine if the
// subject has the required roles according to predefined Casbin policies.
// func (cm *CasbinMiddleware) RequiresRoles(roles []string, opts ...func(o *Options)) echo.HandlerFunc {
func (cm *CasbinMiddleware) RequiresRoles(next echo.HandlerFunc) echo.HandlerFunc {
	roles := []string{"admin"}
	options := &Options{
		ValidationRule:   matchAll,
		PermissionParser: permissionParserWithSeperator(":"),
	}

	return func(c echo.Context) error {
		if len(roles) == 0 {
			return next(c)
		}

		sub := cm.config.Lookup(c)
		// sub := "userid"
		if len(sub) == 0 {
			return cm.config.Unauthorized(c)
		}

		userRoles, err := cm.config.Enforcer.GetRolesForUser(sub)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{})
		}

		if options.ValidationRule == matchAll {
			for _, role := range roles {
				if !utils.ContainsString(userRoles, role) {
					return cm.config.Forbidden(c)
				}
			}
			return next(c)
		} else if options.ValidationRule == atLeastOne {
			for _, role := range roles {
				if utils.ContainsString(userRoles, role) {
					return next(c)
				}
			}
			return cm.config.Forbidden(c)
		}

		return next(c)
	}
}

// ReloadEnforcer ...
func (cm *CasbinMiddleware) ReloadEnforcer(modelFilePath string, adapter *casbinpgadapter.Adapter) {
	mc, _ := model.NewModelFromString(modelFilePath)
	enforcer, err := casbin.NewEnforcer(mc, adapter)
	if err != nil {
		logrus.Errorf("ReloadEnforcer.NewEnforcer():%+v\n", err)
	}
	if err := enforcer.LoadPolicy(); err != nil {
		logrus.Errorf("ReloadEnforcer.LoadPolicy():%+v\n", err)
	}
	cm.config.Enforcer = enforcer
}

func replaceParam(url string, param string) string {
	param = "/" + param
	return strings.Replace(url, param, "", 1)
}

var url2resource = map[string]string{}

var url2cmethod = map[string]string{
	"/api/v1/": "approve",
}

var httpmethod2string = map[string]string{
	"GET":    "list",
	"POST":   "create",
	"PATCH":  "update",
	"PUT":    "update",
	"DELETE": "delete",
}
