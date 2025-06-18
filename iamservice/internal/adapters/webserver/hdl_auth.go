package webserver

import (
	"net/http"
	"strings"

	"github.com/go-openapi/strfmt"
	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase"
	"github.com/mobiletoly/gokatana-samples/iamservice/templates"
)

// AuthWebHandlers handles authentication-related web requests
type AuthWebHandlers struct {
	authUC *usecase.AuthUser
}

// NewAuthWebHandlers creates a new instance of AuthWebHandlers
func NewAuthWebHandlers(authUC *usecase.AuthUser) *AuthWebHandlers {
	return &AuthWebHandlers{
		authUC: authUC,
	}
}

// SignInFormHandler renders the sign-in form
func (h *AuthWebHandlers) SignInFormHandler(c echo.Context) error {
	ctx := c.Request().Context()

	if IsHTMX(c) {
		return templates.SignInForm().Render(ctx, c.Response().Writer)
	}

	// Get authentication status
	userEmail, _ := h.GetAuthenticatedUser(c)

	return templates.Layout("Sign In", templates.SignInForm(), userEmail).Render(ctx, c.Response().Writer)
}

// SignInHandler handles sign-in
func (h *AuthWebHandlers) SignInHandler(c echo.Context) error {
	ctx := c.Request().Context()

	email := strings.TrimSpace(c.FormValue("email"))
	password := strings.TrimSpace(c.FormValue("password"))

	if email == "" || password == "" {
		return templates.SignInError("Email and password are required").Render(ctx, c.Response().Writer)
	}

	// Create signin request
	signinReq := &swagger.SigninRequest{
		Email:    (*strfmt.Email)(&email),
		Password: &password,
	}

	authResp, err := h.authUC.SignIn(ctx, signinReq)
	if err != nil {
		return templates.SignInError("Invalid credentials").Render(ctx, c.Response().Writer)
	}

	// Set secure authentication cookies
	h.setAuthCookies(c, *authResp.AccessToken, *authResp.RefreshToken, email)

	if IsHTMX(c) {
		// For HTMX requests, redirect to home page to refresh the entire layout
		c.Response().Header().Set("HX-Redirect", "/web/admin")
		return c.NoContent(http.StatusOK)
	} else {
		// For regular requests, redirect using standard HTTP redirect
		return c.Redirect(http.StatusSeeOther, "/web/admin")
	}
}

// SignOutHandler handles sign-out
func (h *AuthWebHandlers) SignOutHandler(c echo.Context) error {
	// Clear authentication cookies
	h.clearAuthCookies(c)

	// Redirect to home page
	c.Response().Header().Set("HX-Redirect", "/web/admin")
	return c.NoContent(http.StatusOK)
}

// Helper methods for cookie management

// setAuthCookies sets secure authentication cookies
func (h *AuthWebHandlers) setAuthCookies(c echo.Context, accessToken, refreshToken, email string) {
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
		Secure:   true,  // Set to false for local development over HTTP
		SameSite: http.SameSiteLaxMode,
	}

	c.SetCookie(accessCookie)
	c.SetCookie(refreshCookie)
	c.SetCookie(emailCookie)
}

// clearAuthCookies clears authentication cookies
func (h *AuthWebHandlers) clearAuthCookies(c echo.Context) {
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
func (h *AuthWebHandlers) GetAuthenticatedUser(c echo.Context) (string, bool) {
	// Check if access token exists
	accessCookie, err := c.Cookie("access_token")
	if err != nil {
		return "", false
	}

	// Validate the access token
	_, err = h.authUC.ValidateAccessToken(accessCookie.Value)
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
