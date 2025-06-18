package webserver

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana-samples/iamservice/templates"
)

func renderTemplateComponent(c echo.Context, title string, component templ.Component) error {
	ctx := c.Request().Context()
	if IsHTMX(c) {
		// Return just the content for HTMX requests
		return component.Render(ctx, c.Response().Writer)
	}

	email := ""
	emailCookie, err := c.Cookie("user_email")
	if err == nil {
		email = emailCookie.Value
	}

	return templates.Layout(title, component, email).Render(ctx, c.Response().Writer)
}
