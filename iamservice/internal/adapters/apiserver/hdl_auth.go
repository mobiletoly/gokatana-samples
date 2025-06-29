package apiserver

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/swagger"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/usecase"
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/mobiletoly/gokatana/kathttp_echo"
)

func signupHandler(uc *usecase.AuthMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		var signupReq swagger.SignupRequest
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

func signoutHandler(uc *usecase.AuthMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		// TODO implement it
		return c.JSON(http.StatusOK, struct{}{})
	}
}

func refreshTokenHandler(uc *usecase.AuthMgm) func(c echo.Context) error {
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

func confirmEmailHandler(uc *usecase.AuthMgm) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		userID := c.QueryParam("userId")
		code := c.QueryParam("code")

		if userID == "" || code == "" {
			return kathttp_echo.ReportBadRequest(katapp.NewErr(katapp.ErrInvalidInput, "user ID and confirmation code are required"))
		}

		err := uc.ConfirmEmail(ctx, userID, code)
		if err != nil {
			return kathttp_echo.ReportHTTPError(err)
		}

		response := &swagger.EmailConfirmationResponse{
			Message: "Email confirmed successfully",
		}

		return c.JSON(http.StatusOK, response)
	}
}
