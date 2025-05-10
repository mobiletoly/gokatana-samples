package apiserver_std

import (
	"context"
	"encoding/json"
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/core/usecase"
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/mobiletoly/gokatana/kathttp"
	"github.com/mobiletoly/gokatana/kathttp_std"
	sloghttp "github.com/samber/slog-http"
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
//	-ldflags "-X 'github.com/mobiletoly/gokatana-samples/hexagonal/internal/adapters/apiserver_std.AppTagVersion=1.0.0'"
func init() {
	SampleVersionResponse.Version = AppTagVersion
}

func Start(ctx context.Context, uc *usecase.UseCases) *http.Server {
	logger := katapp.Logger(ctx).Logger
	server := kathttp_std.Start(
		ctx,
		&uc.Config.Server,
		logger,
		func(mux *http.ServeMux) http.Handler {
			apiRoutes(mux, uc)
			config := sloghttp.Config{
				WithRequestID: true,
				WithSpanID:    true,
				WithTraceID:   true,
			}
			handler := sloghttp.NewWithConfig(logger, config)(mux)
			return handler
		})
	return server
}

func WaitForInterruptSignal(ctx context.Context, server *http.Server, timeout time.Duration) {
	katapp.WaitForInterruptSignal(ctx, timeout, func() error {
		return server.Shutdown(ctx)
	})
}

func apiRoutes(mux *http.ServeMux, uc *usecase.UseCases) {
	mux.Handle("GET /api/v1/sample/version", getSampleVersionRoute())
	mux.Handle("GET /api/v1/sample/contacts/{id}", getContactByIDRoute(uc.Contact))
	mux.Handle("GET /api/v1/sample/contacts", getAllContactsRoute(uc.Contact))
	mux.Handle("POST /api/v1/sample/contacts", addContactRoute(uc.Contact))
}

func getSampleVersionRoute() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(SampleVersionResponse); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		katapp.Logger(r.Context()).InfoContext(r.Context(), "version requested")
	}
}
