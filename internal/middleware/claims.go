package middleware

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/o1egl/paseto/v2"
)

type UserClaim struct {
	ID           string `json:"id"`
	DisplayName  string `json:"displayName"`
	PhoneNumber  string `json:"phoneNumber"`
	CountryCode  string `json:"countryCode"`
	RoleID       string `json:"roleID"`
	DepartmentID string `json:"departmentID"`
}

type claimCtxKey int

const (
	_ claimCtxKey = iota
	userClaimKey
)

func WithUserClaim(ctx context.Context, claims UserClaim) context.Context {
	return context.WithValue(ctx, userClaimKey, claims)
}

func UserClaimFromContext(ctx context.Context) UserClaim {
	claims, ok := ctx.Value(userClaimKey).(UserClaim)
	if !ok {
		return UserClaim{}
	}
	return claims
}

func SetClaimsMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, ok := c.Get("user").(paseto.JSONToken)
			if !ok {
				return next(c)
			}
			var user UserClaim
			if err := claims.Get("user", &user); err != nil {
				fmt.Printf("claims.Get():%v\n", err)
				return next(c)
			}
			ctx := c.Request().Context()
			ctx = context.WithValue(ctx, userClaimKey, user)
			c.SetRequest(c.Request().WithContext(ctx))
			fmt.Print("c.SetRequest()")
			return next(c)
		}
	}
}
