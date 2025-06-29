package webserver

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/internal/serverhelp"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/webserver/mw"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/webserver/webadmin"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/webserver/webuser"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase"
	"github.com/mobiletoly/gokatana-samples/iamservice/templates/admin"
	"github.com/mobiletoly/gokatana-samples/iamservice/templates/user"
)

type HomeWebHandlers struct {
	authHandler *webadmin.AuthWebHandlers
}

func NewHomeWebHandlers(authHandler *webadmin.AuthWebHandlers) *HomeWebHandlers {
	return &HomeWebHandlers{
		authHandler: authHandler,
	}
}

// SetupWebRoutes configures all web routes
func SetupWebRoutes(e *echo.Echo, uc *usecase.UseCases) {
	authMiddleware := serverhelp.NewJWTAuthWebServerMiddleware([]byte(uc.Config.Credentials.JwtSecret))

	// Static file serving
	e.Static("/static", "static")

	setupAdminRoutes(e, uc, authMiddleware)
	setupUserRoutes(e, uc, authMiddleware)
}

// setupAdminRoutes wires web interface routes under /web/admin
func setupAdminRoutes(e *echo.Echo, uc *usecase.UseCases, authMiddleware *serverhelp.JWTAuthMiddleware) {
	adminAuthLock := authMiddleware.WithAnyRole("admin", "sysadmin")
	sysadminAuthLock := authMiddleware.WithAnyRole("sysadmin")

	authWeb := webadmin.NewAuthWebHandlers(uc.Auth)
	userMgmWeb := webadmin.NewUserMgmWebHandlers(uc.UserMgm, uc.Auth)
	tenantMgmWeb := webadmin.NewTenantMgmWebHandlers(uc.Auth)

	// Admin web interface routes under /web/admin
	root := e.Group("/web/admin")
	root.Use(mw.RewriteHttpErrorToTemplateMiddleware(func(alert templ.Component, email string) templ.Component {
		return admin.Layout("", alert, email)
	}))

	// Add HTMX middleware to detect HTMX requests
	root.Use(mw.HTMXMiddleware())

	// Main admin routes
	root.GET("", authWeb.HomeLoadHandler)  // /web/admin
	root.GET("/", authWeb.HomeLoadHandler) // /web/admin/

	// User management routes (protected with admin role middleware)
	users := root.Group("/users", adminAuthLock)
	users.GET("", userMgmWeb.UsersListLoadHandler)
	users.GET("/new", userMgmWeb.NewUserLoadHandler)
	users.GET("/:id", userMgmWeb.UserDetailLoadHandler)
	users.GET("/:id/edit", userMgmWeb.UserEditLoadHandler)
	users.GET("/:id/change-password", userMgmWeb.UserChangePasswordLoadHandler)
	users.GET("/:id/roles", userMgmWeb.UserRolesLoadHandler)
	users.POST("", userMgmWeb.CreateUserLoadHandler)
	users.PUT("/:id", userMgmWeb.UpdateUserSubmitHandler)
	users.POST("/:id/change-password", userMgmWeb.ChangePasswordSubmitHandler)
	users.POST("/:id/roles", userMgmWeb.AssignRoleSubmitHandler)
	users.DELETE("/:id", userMgmWeb.DeleteUserSubmitHandler)
	users.DELETE("/:id/roles/:roleName", userMgmWeb.DeleteRoleSubmitHandler)

	// Tenant management routes (protected with sysadmin role middleware)
	tenants := root.Group("/tenants", adminAuthLock)
	tenants.GET("", tenantMgmWeb.TenantsListLoadHandler)
	tenants.GET("/new", tenantMgmWeb.NewTenantLoadHandler, sysadminAuthLock)
	tenants.GET("/:id", tenantMgmWeb.TenantDetailLoadHandler)
	tenants.GET("/:id/edit", tenantMgmWeb.TenantEditLoadHandler)
	tenants.POST("", tenantMgmWeb.CreateTenantSubmitHandler)
	tenants.PUT("/:id", tenantMgmWeb.UpdateTenantSubmitHandler)
	tenants.DELETE("/:id", tenantMgmWeb.DeleteTenantSubmitHandler)

	// Authentication routes
	auth := root.Group("/auth")
	auth.GET("/signin", authWeb.SignInLoadHandler)
	auth.POST("/signin", authWeb.SignInSubmitHandler)
	auth.POST("/signout", authWeb.SignOutSubmitHandler)
}

// setupUserRoutes wires web interface routes under /web/user
func setupUserRoutes(e *echo.Echo, uc *usecase.UseCases, authMiddleware *serverhelp.JWTAuthMiddleware) {
	authLock := authMiddleware.WithAnyRole("admin", "sysadmin", "user")

	authWeb := webuser.NewAuthWebHandlers(uc.Auth)
	accountWeb := webuser.NewAccountWebHandlers(uc.Auth, uc.UserMgm, uc.UserProfileMgm)

	root := e.Group("/web/user")
	root.Use(mw.RewriteHttpErrorToTemplateMiddleware(func(alert templ.Component, email string) templ.Component {
		return user.Layout("", alert, email)
	}))
	root.Use(mw.HTMXMiddleware())

	// Main user routes
	root.GET("", authWeb.HomeLoadHandler)  // /web/user
	root.GET("/", authWeb.HomeLoadHandler) // /web/user/

	// User authentication routes
	auth := root.Group("/auth")
	auth.GET("/signin", authWeb.SignInLoadHandler)
	auth.POST("/signin", authWeb.SignInSubmitHandler)
	auth.GET("/signup", authWeb.SignUpLoadHandler)
	auth.POST("/signup", authWeb.SignUpSubmitHandler)
	auth.POST("/signout", authWeb.SignOutSubmitHandler)

	// User account routes (protected)
	account := root.Group("/account", authLock)
	account.GET("", accountWeb.AccountLoadHandler)
	account.GET("/edit", accountWeb.EditAccountLoadHandler)
	account.PUT("/update", accountWeb.UpdateAccountSubmitHandler)
	account.GET("/change-password", accountWeb.ChangePasswordLoadHandler)
	account.PUT("/change-password", accountWeb.UpdatePasswordSubmitHandler)

	// User profile routes (protected)
	profile := root.Group("/profile", authLock)
	profile.GET("/edit", accountWeb.EditProfileLoadHandler, authLock)
	profile.PUT("/update", accountWeb.UpdateProfileSubmitHandler, authLock)
}
