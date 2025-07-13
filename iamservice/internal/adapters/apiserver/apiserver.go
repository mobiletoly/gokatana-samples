package apiserver

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/internal/serverhelp"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/webserver"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase"
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/mobiletoly/gokatana/kathttp"
	"github.com/mobiletoly/gokatana/kathttp_echo"
	"github.com/samber/slog-echo"
	"net/http"
	"time"
)

var AppTagVersion = "undefined"

var HttpVersionResponse = kathttp.Version{
	Healthy: true,
	Service: "iamservice",
}

// You can set the version during compile time by passing ldflags to the go build command:
//
//	-ldflags "-X 'github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/apiserver_echo.AppTagVersion=1.0.0'"
func init() {
	HttpVersionResponse.Version = AppTagVersion
}

func Start(ctx context.Context, uc *usecase.UseCases) *echo.Echo {
	logger := katapp.Logger(ctx).Logger
	server := kathttp_echo.Start(
		ctx,
		&uc.Config.Server,
		logger,
		func(e *echo.Echo) {
			config := slogecho.Config{
				WithRequestID: true,
				WithSpanID:    true,
				WithTraceID:   true,
			}
			e.Use(slogecho.NewWithConfig(logger, config))
			apiRoutes(e, uc)
			webserver.SetupWebRoutes(e, uc)
		})
	return server
}

// WaitForInterruptSignal waits for interrupt signal to gracefully shut down the server with a timeout.
func WaitForInterruptSignal(ctx context.Context, server *echo.Echo, timeout time.Duration) {
	katapp.WaitForInterruptSignal(ctx, timeout, func() error {
		return server.Shutdown(ctx)
	})
}

func apiRoutes(e *echo.Echo, uc *usecase.UseCases) {
	authMiddleware := serverhelp.NewJWTAuthApiServerMiddleware([]byte(uc.Config.Credentials.JwtSecret))
	authLock := authMiddleware.WithAnyRole("admin", "sysadmin", "user")
	adminAuthLock := authMiddleware.WithAnyRole("admin", "sysadmin")
	sysadminAuthLock := authMiddleware.WithAnyRole("sysadmin")

	api := e.Group("/api/v1")
	api.GET("/version", getHttpVersionRoute())

	// Authentication routes
	auth := api.Group("/auth")
	auth.POST("/signup", signupHandler(uc.Auth))
	auth.POST("/signin", signinHandler(uc.Auth))
	auth.POST("/signout", signoutHandler(uc.Auth), authLock)
	auth.POST("/refresh", refreshTokenHandler(uc.Auth))
	auth.POST("/confirm-email", confirmEmailHandler(uc.Auth))

	// User profile routes (basic authentication required)
	api.GET("/users/me", getMyUserHandler(uc.UserMgm), authLock)

	// User Management API routes (sysadmin role required)
	api.GET("/users/all", listAllUsersHandler(uc.UserMgm), sysadminAuthLock) // GET /api/v1/users/all

	// User Management API routes (admin role required)
	users := api.Group("/users", authLock)
	users.GET("", listAllUsersByTenantHandler(uc.UserMgm))                                     // GET /api/v1/users
	users.GET("/:userId", getUserByIdHandler(uc.UserMgm))                                      // GET /api/v1/users/{userId}
	users.PUT("/:userId", updateAuthUserHandler(uc.UserMgm))                                   // PUT /api/v1/users/{userId}
	users.GET("/:userId/profile", getUserProfileHandler(uc.UserProfileMgm))                    // GET /api/v1/users/{userId}/profile
	users.PUT("/:userId/profile", updateUserProfileHandler(uc.UserProfileMgm))                 // PUT /api/v1/users/{userId}/profile
	users.GET("/:userId/roles", getUserRolesHandler(uc.UserMgm))                               // GET /api/v1/users/{userId}/roles
	users.POST("/:userId/roles", assignUserRoleHandler(uc.UserMgm), adminAuthLock)             // POST /api/v1/users/{userId}/roles
	users.DELETE("/:userId/roles/:roleName", deleteUserRoleHandler(uc.UserMgm), adminAuthLock) // DELETE /api/v1/users/{userId}/roles/{roleName}

	// Tenant Management API routes (sysadmin role required)
	tenants := api.Group("/tenants", authLock)
	tenants.GET("", getAllTenantsHandler(uc.Auth))                               // GET /api/v1/tenants
	tenants.POST("", createTenantHandler(uc.Auth), sysadminAuthLock)             // POST /api/v1/tenants
	tenants.GET("/:tenantId", getTenantByIdHandler(uc.Auth))                     // GET /api/v1/tenants/{tenantId}
	tenants.PUT("/:tenantId", updateTenantHandler(uc.Auth), adminAuthLock)       // PUT /api/v1/tenants/{tenantId}
	tenants.DELETE("/:tenantId", deleteTenantHandler(uc.Auth), sysadminAuthLock) // DELETE /api/v1/tenants/{tenantId}
}

func getHttpVersionRoute() func(c echo.Context) error {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, HttpVersionResponse)
	}
}
