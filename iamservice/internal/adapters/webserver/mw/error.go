package mw

import (
	"errors"
	"fmt"
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana-samples/iamservice/templates/common"
	"github.com/mobiletoly/gokatana/kathttp"
	"github.com/mobiletoly/gokatana/kathttp_echo"
	"net/http"
)

func RewriteHttpErrorToTemplateMiddleware(component func(alert templ.Component, email string) templ.Component) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		fn := func(c echo.Context) error {
			reqErr := next(c)
			if reqErr == nil {
				return nil
			}

			var he *echo.HTTPError
			if !errors.As(reqErr, &he) {
				he = kathttp_echo.ReportHTTPError(reqErr)
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

			var msg string
			if he.Code == http.StatusBadRequest {
				msg = fmt.Sprintf("%v", details)
			} else {
				msg = fmt.Sprintf("%s: %v", status, details)
			}

			alert := common.ErrorAlert(msg)
			if IsHTMX(c) {
				return alert.Render(ctx, c.Response().Writer)
			}
			return component(alert, email).Render(ctx, c.Response().Writer)
		}
		return fn
	}
}
