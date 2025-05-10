package apiserver_echo

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/core/usecase"
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/mobiletoly/gokatana/kathttp"
	"github.com/mobiletoly/gokatana/kathttp_echo"
	"github.com/samber/slog-echo"
	"net/http"
	"time"
)

var AppTagVersion = "undefined"

var SampleVersionResponse = kathttp.Version{
	Healthy: true,
	Service: "sample-hexagonal",
}

// You can set the version during compile time by passing ldflags to the go build command:
//
//	-ldflags "-X 'github.com/mobiletoly/gokatana-samples/hexagonal/internal/adapters/apiserver_echo.AppTagVersion=1.0.0'"
func init() {
	SampleVersionResponse.Version = AppTagVersion
}

func Start(ctx context.Context, uc *usecase.UseCases) *echo.Echo {
	logger := katapp.Logger(ctx).Logger
	server := kathttp_echo.Start(
		ctx,
		&uc.Config.Server,
		logger,
		func(e *echo.Echo) {
			config := slogecho.Config{
				WithRequestID: true,
				WithSpanID:    true,
				WithTraceID:   true,
			}
			e.Use(slogecho.NewWithConfig(logger, config))
			apiRoutes(e, uc)
		})
	return server
}

// WaitForInterruptSignal waits for interrupt signal to gracefully shut down the server with a timeout.
func WaitForInterruptSignal(ctx context.Context, server *echo.Echo, timeout time.Duration) {
	katapp.WaitForInterruptSignal(ctx, timeout, func() error {
		return server.Shutdown(ctx)
	})
}

func apiRoutes(e *echo.Echo, uc *usecase.UseCases) {
	api := e.Group("/api/v1")
	sample := api.Group("/sample")
	sample.GET("/version", getSampleVersionRoute())

	contacts := sample.Group("/contacts")
	contacts.GET("/:id", getContactByIDRoute(uc.Contact))
	contacts.GET("", getAllContactsRoute(uc.Contact))
	contacts.POST("", addContactRoute(uc.Contact))
}

func getSampleVersionRoute() func(c echo.Context) error {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, SampleVersionResponse)
	}
}
