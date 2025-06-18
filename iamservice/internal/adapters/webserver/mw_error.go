package webserver

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana-samples/iamservice/templates"
	"github.com/mobiletoly/gokatana/kathttp"
	"net/http"
)

func rewriteHttpErrorToTemplateMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		reqErr := next(c)
		if reqErr == nil {
			return nil
		}

		var he *echo.HTTPError
		if !errors.As(reqErr, &he) {
			return reqErr
		}

		ctx := c.Request().Context()
		email := ""
		emailCookie, err := c.Cookie("user_email")
		if err == nil {
			email = emailCookie.Value
		}

		status := http.StatusText(he.Code)
		details := he.Message
		if errResp, ok := he.Message.(*kathttp.ErrResponse); ok {
			details = errResp.ErrorText
		}

		msg := fmt.Sprintf("%s: %v", status, details)
		if IsHTMX(c) {
			return templates.UserFormError(msg).Render(ctx, c.Response().Writer)
		}
		return templates.
			Layout("", templates.UserFormError(msg), email).
			Render(ctx, c.Response().Writer)
	}
}
