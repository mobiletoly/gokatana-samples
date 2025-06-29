package mw

import (
	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase"
	"net/http"
	"strings"
)

// SetAuthCookies sets secure authentication cookies
func SetAuthCookies(c echo.Context, accessToken, refreshToken, email string) {
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

// ClearAuthCookies clears authentication cookies
func ClearAuthCookies(c echo.Context) {
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

// GetAuthenticatedUserEmailFromCookie returns the authenticated user's email from cookies
func GetAuthenticatedUserEmailFromCookie(c echo.Context, auth *usecase.AuthMgm) (string, bool) {
	// Check if access token exists
	accessCookie, err := c.Cookie("access_token")
	if err != nil {
		return "", false
	}

	// Validate the access token
	_, err = auth.ValidateAccessToken(accessCookie.Value)
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
