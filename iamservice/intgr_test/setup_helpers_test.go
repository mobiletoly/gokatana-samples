package intgr_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	// Pick one of the imports below, depending on which framework you use
	// -- Option 1: Echo framework
	apiserver "github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/apiserver"
	// -- Option 2: Chi router
	// apiserver "github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/apiserver_chi"
	// -- Option 3: Standard net/http
	//apiserver "github.com/mobiletoly/gokatana-samples/iamservice/internal/adapters/apiserver_std"

	"github.com/mobiletoly/gokatana-samples/iamservice/internal/core/app"
	"github.com/mobiletoly/gokatana-samples/iamservice/internal/infra"
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/mobiletoly/gokatana/kathttpc"
	"github.com/mobiletoly/gokatana/katpg"
)

// TestEnvironment holds the test environment setup
type TestEnvironment struct {
	Context   context.Context
	AppConfig *app.Config
	T         *testing.T
}

// SetupTestEnvironment initializes the test environment with database and server
func SetupTestEnvironment(t *testing.T) *TestEnvironment {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx := katapp.ContextWithAppLogger(logger)
	ctx = katapp.ContextWithRunInTest(ctx, true)

	dbMigrate := "../dbmigrate"
	pc := katpg.RunPostgresTestContainer(ctx, t, &dbMigrate, []string{
		"init/sample_data.sql",
	})
	t.Cleanup(func() {
		pc.Terminate(ctx, t)
	})

	started := make(chan struct{})
	var appConfig *app.Config
	go func() {
		infra.Start("test",
			func(cfg *app.Config) {
				pc.ApplyToConfig(&cfg.Database)
				appConfig = cfg
			},
			started,
		)
	}()
	<-started
	kathttpc.WaitForURLToBecomeReady(ctx, kathttpc.LocalURL(appConfig.Server.Port, "api/v1/version"))

	return &TestEnvironment{
		Context:   ctx,
		AppConfig: appConfig,
		T:         t,
	}
}

// GetAPIServerInfo returns API server information for tests
func GetAPIServerInfo() (string, string, string) {
	return apiserver.HttpVersionResponse.Service, apiserver.AppTagVersion, "true"
}
