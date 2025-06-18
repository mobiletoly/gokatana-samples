package webserver

import (
	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/internal/serverhelp"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase"
	"github.com/mobiletoly/gokatana-samples/iamservice/templates"
)

type HomeWebHandlers struct {
	authHandler *AuthWebHandlers
}

func NewHomeWebHandlers(authHandler *AuthWebHandlers) *HomeWebHandlers {
	return &HomeWebHandlers{
		authHandler: authHandler,
	}
}

// SetupWebRoutes configures all web routes
func SetupWebRoutes(e *echo.Echo, uc *usecase.UseCases) {
	authMiddleware := serverhelp.NewJWTAuthWebServerMiddleware(uc.Auth.GetJWTSecret())
	adminAuthLock := authMiddleware.WithAnyRole("admin")

	// Create handler instances
	authHandlers := NewAuthWebHandlers(uc.Auth)
	homeHandlers := NewHomeWebHandlers(authHandlers)
	userMgmHandlers := NewUserMgmWebHandlers(uc.UserMgm, uc.Auth, authHandlers)

	// Static file serving
	e.Static("/static", "static")

	// Admin web interface routes under /web/admin
	admin := e.Group("/web/admin")
	admin.Use(rewriteHttpErrorToTemplateMiddleware)

	// Add HTMX middleware to detect HTMX requests
	admin.Use(HTMXMiddleware())

	// Main admin routes
	admin.GET("", homeHandlers.HomeHandler)  // /web/admin
	admin.GET("/", homeHandlers.HomeHandler) // /web/admin/

	// User management routes (protected with admin role middleware)
	users := admin.Group("/users", adminAuthLock)
	users.GET("", userMgmHandlers.UsersListHandler)
	users.GET("/new", userMgmHandlers.UserFormHandler)
	users.GET("/:id", userMgmHandlers.UserDetailHandler)
	users.GET("/:id/roles", userMgmHandlers.UserRolesHandler)
	users.POST("", userMgmHandlers.CreateUserHandler)
	users.POST("/:id/roles", userMgmHandlers.AssignRoleHandler)
	users.DELETE("/:id", userMgmHandlers.DeleteUserHandler)
	users.DELETE("/:id/roles/:roleName", userMgmHandlers.DeleteRoleHandler)

	// Authentication routes
	admin.GET("/auth/signin", authHandlers.SignInFormHandler)
	admin.POST("/auth/signin", authHandlers.SignInHandler)
	admin.POST("/auth/signout", authHandlers.SignOutHandler)
}

// HomeHandler renders the home page
func (h *HomeWebHandlers) HomeHandler(c echo.Context) error {
	return renderTemplateComponent(c, "Home", templates.Home())
}
