package webuser

import (
	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/webserver/mw"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase"
	"github.com/mobiletoly/gokatana-samples/iamservice/templates/user"
	"net/http"
	"strings"
)

type AuthWebHandlers struct {
	authMgm *usecase.AuthMgm
}

func NewAuthWebHandlers(authMgm *usecase.AuthMgm) *AuthWebHandlers {
	return &AuthWebHandlers{
		authMgm: authMgm,
	}
}

// HomeLoadHandler renders the user dashboard home page
func (a *AuthWebHandlers) HomeLoadHandler(c echo.Context) error {
	userEmail, _ := mw.GetAuthenticatedUserEmailFromCookie(c, a.authMgm)
	return renderTemplateComponent(c, "Dashboard", user.Home(userEmail))
}

// SignInLoadHandler renders the sign-in form
func (a *AuthWebHandlers) SignInLoadHandler(c echo.Context) error {
	ctx := c.Request().Context()
	if mw.IsHTMX(c) {
		return user.SignInForm().Render(ctx, c.Response().Writer)
	}
	userEmail, _ := a.GetAuthenticatedUser(c)
	return user.Layout("Sign In", user.SignInForm(), userEmail).Render(ctx, c.Response().Writer)
}

// SignUpLoadHandler renders the sign-up form
func (a *AuthWebHandlers) SignUpLoadHandler(c echo.Context) error {
	ctx := c.Request().Context()
	if mw.IsHTMX(c) {
		return user.SignUpForm().Render(ctx, c.Response().Writer)
	}
	userEmail, _ := a.GetAuthenticatedUser(c)
	return user.Layout("Sign Up", user.SignUpForm(), userEmail).Render(ctx, c.Response().Writer)
}

// SignInSubmitHandler handles sign-in
func (a *AuthWebHandlers) SignInSubmitHandler(c echo.Context) error {
	ctx := c.Request().Context()

	tenantId := strings.TrimSpace(c.FormValue("tenantId"))
	email := strings.TrimSpace(c.FormValue("email"))
	password := strings.TrimSpace(c.FormValue("password"))
	signinReq := &swagger.SigninRequest{
		Email:    email,
		Password: password,
		TenantId: tenantId,
	}
	authResp, err := a.authMgm.SignIn(ctx, signinReq)
	if err != nil {
		return err
	}

	a.setAuthCookies(c, authResp.AccessToken, authResp.RefreshToken, email)

	if mw.IsHTMX(c) {
		// For HTMX requests, redirect to home page to refresh the entire layout
		c.Response().Header().Set("HX-Redirect", "/web/user")
		return c.NoContent(http.StatusOK)
	} else {
		// For regular requests, redirect using standard HTTP redirect
		return c.Redirect(http.StatusSeeOther, "/web/user")
	}
}

// SignUpSubmitHandler handles sign-up
func (a *AuthWebHandlers) SignUpSubmitHandler(c echo.Context) error {
	ctx := c.Request().Context()

	tenantId := strings.TrimSpace(c.FormValue("tenantId"))
	firstName := strings.TrimSpace(c.FormValue("firstName"))
	lastName := strings.TrimSpace(c.FormValue("lastName"))
	email := strings.TrimSpace(c.FormValue("email"))
	password := strings.TrimSpace(c.FormValue("password"))

	signupReq := &swagger.SignupRequest{
		TenantId:  tenantId,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Password:  password,
		Source:    "web",
	}
	_, err := a.authMgm.SignUp(ctx, signupReq)
	if err != nil {
		return err
	}
	return user.SignUpSuccess().Render(ctx, c.Response().Writer)
}

// SignOutSubmitHandler handles sign-out
func (a *AuthWebHandlers) SignOutSubmitHandler(c echo.Context) error {
	a.clearAuthCookies(c)

	// Redirect to home page
	c.Response().Header().Set("HX-Redirect", "/web/user")
	return c.NoContent(http.StatusOK)
}

// setAuthCookies sets secure authentication cookies
func (a *AuthWebHandlers) setAuthCookies(c echo.Context, accessToken, refreshToken, email string) {
	// Set Secure flag based on environment
	isLocalDev := strings.Contains(c.Request().Host, "localhost") ||
		strings.Contains(c.Request().Host, "127.0.0.1")
	isHTTPS := c.Request().Header.Get("X-Forwarded-Proto") == "https" ||
		c.Request().TLS != nil
	secureCookie := false
	if !isLocalDev && isHTTPS {
		secureCookie = true
	}

	// Access token cookie (shorter expiry)
	accessCookie := &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		MaxAge:   3600, // 1 hour
		HttpOnly: true,
		Secure:   secureCookie,
		SameSite: http.SameSiteLaxMode,
	}

	// Refresh token cookie (longer expiry)
	refreshCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   secureCookie,
		SameSite: http.SameSiteLaxMode,
	}

	// User email cookie (for display purposes, not sensitive)
	emailCookie := &http.Cookie{
		Name:     "user_email",
		Value:    email,
		Path:     "/",
		MaxAge:   3600,  // 1 hour
		HttpOnly: false, // Allow JavaScript access for display
		Secure:   secureCookie,
		SameSite: http.SameSiteLaxMode,
	}

	c.SetCookie(accessCookie)
	c.SetCookie(refreshCookie)
	c.SetCookie(emailCookie)
}

// clearAuthCookies clears authentication cookies
func (a *AuthWebHandlers) clearAuthCookies(c echo.Context) {
	cookies := []string{"access_token", "refresh_token", "user_email"}

	for _, name := range cookies {
		cookie := &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1, // Delete cookie
			HttpOnly: true,
			Secure:   false, // Set based on environment
			SameSite: http.SameSiteLaxMode,
		}
		c.SetCookie(cookie)
	}
}

// GetAuthenticatedUser returns the authenticated user's email from cookies
func (a *AuthWebHandlers) GetAuthenticatedUser(c echo.Context) (string, bool) {
	// Check if access token exists
	accessCookie, err := c.Cookie("access_token")
	if err != nil {
		return "", false
	}

	// Validate the access token
	_, err = a.authMgm.ValidateAccessToken(accessCookie.Value)
	if err != nil {
		return "", false
	}

	// Get user email from cookie
	emailCookie, err := c.Cookie("user_email")
	if err != nil {
		return "", false
	}

	return emailCookie.Value, true
}
