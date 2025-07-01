package mw

import (
	"github.com/labstack/echo/v4"
)

// HTMXMiddleware detects HTMX requests and stores the result in context
func HTMXMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if this is an HTMX request
			isHTMX := c.Request().Header.Get("HX-Request") == "true"

			// Store in context for easy access in handlers
			c.Set("isHTMX", isHTMX)
			return next(c)
		}
	}
}

// IsHTMX returns whether the current request is an HTMX request
func IsHTMX(c echo.Context) bool {
	if val := c.Get("isHTMX"); val != nil {
		if isHTMX, ok := val.(bool); ok {
			return isHTMX
		}
	}
	return false
}
