package webadmin

import (
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/webserver/mw"
	"github.com/mobiletoly/gokatana-samples/iamservice/templates/admin"
	"net/http"
	"strings"

	"github.com/go-openapi/strfmt"
	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase"
)

// AuthWebHandlers handles authentication-related web requests
type AuthWebHandlers struct {
	authMgm *usecase.AuthMgm
}

// HomeLoadHandler renders the home page
func (h *AuthWebHandlers) HomeLoadHandler(c echo.Context) error {
	return renderTemplateComponent(c, "Home", admin.Home())
}

// NewAuthWebHandlers creates a new instance of AuthWebHandlers
func NewAuthWebHandlers(authUC *usecase.AuthMgm) *AuthWebHandlers {
	return &AuthWebHandlers{
		authMgm: authUC,
	}
}

// SignInLoadHandler renders the sign-in form
func (h *AuthWebHandlers) SignInLoadHandler(c echo.Context) error {
	ctx := c.Request().Context()
	if mw.IsHTMX(c) {
		return admin.SignInForm().Render(ctx, c.Response().Writer)
	}
	userEmail, _ := mw.GetAuthenticatedUserEmailFromCookie(c, h.authMgm)
	return admin.Layout("Sign In", admin.SignInForm(), userEmail).Render(ctx, c.Response().Writer)
}

// SignInSubmitHandler handles sign-in
func (h *AuthWebHandlers) SignInSubmitHandler(c echo.Context) error {
	ctx := c.Request().Context()

	tenantId := strings.TrimSpace(c.FormValue("tenantId"))
	email := strings.TrimSpace(c.FormValue("email"))
	password := strings.TrimSpace(c.FormValue("password"))
	signinReq := &swagger.SigninRequest{
		Email:    strfmt.Email(email),
		Password: password,
		TenantID: tenantId,
	}

	authResp, err := h.authMgm.SignIn(ctx, signinReq)
	if err != nil {
		return err
	}
	mw.SetAuthCookies(c, authResp.AccessToken, authResp.RefreshToken, email)
	if mw.IsHTMX(c) {
		// For HTMX requests, redirect to home page to refresh the entire layout
		c.Response().Header().Set("HX-Redirect", "/web/admin")
		return c.NoContent(http.StatusOK)
	} else {
		// For regular requests, redirect using standard HTTP redirect
		return c.Redirect(http.StatusSeeOther, "/web/admin")
	}
}

// SignOutSubmitHandler handles sign-out
func (h *AuthWebHandlers) SignOutSubmitHandler(c echo.Context) error {
	// Clear authentication cookies
	mw.ClearAuthCookies(c)

	// Redirect to home page
	c.Response().Header().Set("HX-Redirect", "/web/admin")
	return c.NoContent(http.StatusOK)
}
