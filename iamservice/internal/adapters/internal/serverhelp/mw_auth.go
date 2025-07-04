package serverhelp

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase"
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/mobiletoly/gokatana/kathttp_echo"
	"github.com/samber/lo"
	"net/http"
)

type jwtAuthUserClaims struct {
	Roles    []string `json:"roles"`
	TenantID string   `json:"tenantId"`
	jwt.RegisteredClaims
}

type JWTAuthMiddleware struct {
	adminJwtConfig *echojwt.Config
}

func NewJWTAuthApiServerMiddleware(jwtSecret []byte) JWTAuthMiddleware {
	adminJwtConfig := echojwt.Config{
		SigningKey: jwtSecret,
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(jwtAuthUserClaims)
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return kathttp_echo.ReportUnauthorized(err)
		},
	}
	return JWTAuthMiddleware{
		adminJwtConfig: &adminJwtConfig,
	}
}

func NewJWTAuthWebServerMiddleware(jwtSecret []byte) *JWTAuthMiddleware {
	adminJwtConfig := echojwt.Config{
		SigningKey: jwtSecret,
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(jwtAuthUserClaims)
		},
		TokenLookup: "cookie:access_token",
		ErrorHandler: func(c echo.Context, err error) error {
			return echo.NewHTTPError(http.StatusUnauthorized, "You must sign in to access this page")
		},
	}
	return &JWTAuthMiddleware{
		adminJwtConfig: &adminJwtConfig,
	}
}

// protectWithRoles returns a middleware that enforces at least one of the given roles.
func protectWithRoles(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			u := c.Get("user") // comes from echo-jwt
			if u == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing or invalid token")
			}
			token, ok := u.(*jwt.Token)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token type")
			}
			claims, ok := token.Claims.(*jwtAuthUserClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusForbidden, "invalid JWT claims")
			}
			// allow if any requested role is in the user's claims
			if len(roles) == 0 {
				return next(c)
			}
			for _, reqRole := range roles {
				if lo.Contains(claims.Roles, reqRole) {
					return next(c)
				}
			}
			return echo.NewHTTPError(http.StatusForbidden, "access denied: insufficient role")
		}
	}
}

// WithAnyRole returns a single middleware that first applies JWT parsing
// then enforces at least one of the given roles.
func (j JWTAuthMiddleware) WithAnyRole(roles ...string) echo.MiddlewareFunc {
	jwtMw := echojwt.WithConfig(*j.adminJwtConfig)
	protect := protectWithRoles(roles...)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		// First enforce the roles after JWT has run
		handler := protect(next)
		// Then apply JWT parsing so that c.Get("user") is populated before protect
		return jwtMw(handler)
	}
}

// GetUserPrincipalFromToken extracts UserPrincipal from JWT token
func GetUserPrincipalFromToken(c echo.Context) (*usecase.UserPrincipal, error) {
	token, ok := c.Get("user").(*jwt.Token)
	if !ok || token == nil {
		msg := "failed to get user principal from token: missing token"
		katapp.Logger(c.Request().Context()).Error(msg)
		return nil, kathttp_echo.ReportUnauthorized(errors.New(msg))
	}

	claims, ok := token.Claims.(*jwtAuthUserClaims)
	if !ok {
		msg := "failed to get user principal from token: invalid claims"
		katapp.Logger(c.Request().Context()).Error(msg)
		return nil, kathttp_echo.ReportUnauthorized(errors.New(msg))
	}

	// Extract user ID from Subject claim
	userID := claims.Subject
	if userID == "" {
		msg := "failed to get user principal from token: missing user ID"
		katapp.Logger(c.Request().Context()).Error(msg)
		return nil, kathttp_echo.ReportUnauthorized(errors.New(msg))
	}

	// Extract email from Issuer claim (if available) or leave empty
	email := claims.Issuer // This might need to be adjusted based on your JWT structure

	return &usecase.UserPrincipal{
		UserID:   userID,
		TenantID: claims.TenantID,
		Email:    email,
		Roles:    claims.Roles,
	}, nil
}
