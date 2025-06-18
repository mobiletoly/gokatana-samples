package apiserver

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase"
	"github.com/mobiletoly/gokatana/kathttp_echo"
)

// signupHandler handles user registration
func signupHandler(uc *usecase.AuthUser) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		var signupReq swagger.SignupRequest
		if err := c.Bind(&signupReq); err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		authResponse, err := uc.SignUp(ctx, &signupReq)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		return c.JSON(http.StatusCreated, authResponse)
	}
}

// signinHandler handles user authentication
func signinHandler(uc *usecase.AuthUser) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		var signinReq swagger.SigninRequest
		if err := c.Bind(&signinReq); err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		authResponse, err := uc.SignIn(ctx, &signinReq)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		return c.JSON(http.StatusOK, authResponse)
	}
}

// signoutHandler handles user sign out
func signoutHandler(uc *usecase.AuthUser) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		// Get refresh token from request body or header
		refreshToken := c.Request().Header.Get("X-Refresh-Token")
		if refreshToken == "" {
			// Try to get from request body
			var body map[string]string
			if err := c.Bind(&body); err == nil {
				refreshToken = body["refreshToken"]
			}
		}

		messageResponse, err := uc.SignOut(ctx, refreshToken)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		return c.JSON(http.StatusOK, messageResponse)
	}
}

// refreshTokenHandler handles token refresh
func refreshTokenHandler(uc *usecase.AuthUser) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		var refreshReq swagger.RefreshRequest
		if err := c.Bind(&refreshReq); err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		authResponse, err := uc.RefreshToken(ctx, &refreshReq)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		return c.JSON(http.StatusOK, authResponse)
	}
}
