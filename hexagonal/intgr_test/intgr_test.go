package intgr_test

import (
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/adapters/apiserver"
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/core/app"
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/core/model"
	"github.com/mobiletoly/gokatana-samples/hexagonal/internal/infra"
	"github.com/mobiletoly/gokatana/katapp"
	"github.com/mobiletoly/gokatana/kathttp"
	"github.com/mobiletoly/gokatana/kathttpc"
	"github.com/mobiletoly/gokatana/katpg"
	"github.com/mobiletoly/gokatana/kattest"
	"log/slog"
	"os"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestAPIRoutes(t *testing.T) {
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
	kathttpc.WaitForURLToBecomeReady(ctx, &appConfig.Server, "api/v1/sample/version")

	t.Run("API Routes", func(t *testing.T) {
		t.Run("GET /version must succeed", func(t *testing.T) {
			resp, _, err := kattest.LocalHttpJsonGetRequest[kathttp.Version](ctx, &appConfig.Server,
				"api/v1/sample/version", nil)
			assert.NoError(t, err)
			assert.Equal(t, apiserver.SampleVersionResponse.Service, resp.Service)
			assert.Equal(t, true, resp.Healthy)
			assert.Equal(t, apiserver.AppTagVersion, resp.Version)
		})
		t.Run("GET /contacts", func(t *testing.T) {
			t.Run("must succeed", func(t *testing.T) {
				resp, _, err := kattest.LocalHttpJsonGetRequest[[]model.Contact](ctx, &appConfig.Server,
					"api/v1/sample/contacts", map[string]string{})
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, len(*resp), 6)
				item := (*resp)[0]
				assert.EqualValues(t, lo.ToPtr("1"), item.ID)
				assert.EqualValues(t, lo.ToPtr("John"), item.FirstName)
				assert.EqualValues(t, lo.ToPtr("Doe"), item.LastName)
			})
		})
		t.Run("GET /contacts/{id}", func(t *testing.T) {
			t.Run("must succeed", func(t *testing.T) {
				item, _, err := kattest.LocalHttpJsonGetRequest[model.Contact](ctx, &appConfig.Server,
					"api/v1/sample/contacts/1", map[string]string{})
				assert.NoError(t, err)
				assert.EqualValues(t, lo.ToPtr("1"), item.ID)
				assert.EqualValues(t, lo.ToPtr("John"), item.FirstName)
				assert.EqualValues(t, lo.ToPtr("Doe"), item.LastName)
			})
		})
		t.Run("POST /contacts", func(t *testing.T) {
			addContact := &model.AddContact{
				FirstName: lo.ToPtr("Joe"),
				LastName:  lo.ToPtr("Doe"),
			}
			t.Run("must succeed", func(t *testing.T) {
				item, _, err := kattest.LocalHttpJsonPostRequest[model.AddContact, model.Contact](
					ctx, &appConfig.Server, "api/v1/sample/contacts", map[string]string{}, addContact)
				assert.NoError(t, err)
				assert.NotNil(t, item)
				assert.NotEmpty(t, item.ID)
				assert.EqualValues(t, "Joe", *item.FirstName)
				assert.EqualValues(t, "Doe", *item.LastName)
			})
		})
	})
}
