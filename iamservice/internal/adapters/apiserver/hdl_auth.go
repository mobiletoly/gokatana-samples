package apiserver

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase"
	"github.com/mobiletoly/gokatana/kathttp_echo"
)

func signupHandler(uc *usecase.AuthMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		var signupReq swagger.SignUpRequest
		if err := c.Bind(&signupReq); err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		signupResponse, err := uc.SignUp(ctx, &signupReq)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		return c.JSON(http.StatusCreated, signupResponse)
	}
}

func signinHandler(uc *usecase.AuthMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		var signinReq swagger.SignInRequest
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

func signoutHandler(uc *usecase.AuthMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		// Get user ID from the JWT token in the Authorization header
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return kathttp_echo.ReportBadRequest(errors.New("authorization header is required"))
		}

		userID, err := uc.ValidateAccessToken(authHeader)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		// Revoke all refresh tokens for the user
		err = uc.SignOut(ctx, userID)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Successfully signed out"})
	}
}

func refreshTokenHandler(uc *usecase.AuthMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		var refreshReq swagger.TokenRefreshRequest
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

func confirmEmailHandler(uc *usecase.AuthMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		var confirmReq swagger.EmailConfirmationRequest
		if err := c.Bind(&confirmReq); err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		err := uc.ConfirmEmail(ctx, confirmReq.UserId, confirmReq.Code)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		response := &swagger.EmailConfirmationResponse{
			Message: "Email confirmed successfully",
		}

		return c.JSON(http.StatusOK, response)
	}
}
