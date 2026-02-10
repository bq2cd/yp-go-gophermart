package integration

import (
	"context"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/app"
)

type StopFunc func() error

type TestContext struct {
	APIContext *APITestContext
	httpServer *app.HTTPServer
}

func (c *TestContext) Start(baseCtx context.Context) StopFunc {
	ctx, cancel := context.WithCancel(baseCtx)

	grp := new(errgroup.Group)

	grp.Go(func() error {
		return c.httpServer.Run(ctx)
	})

	return func() error {
		cancel()

		c.APIContext.accrualServer.Close()

		return grp.Wait()
	}
}

func (c *TestContext) IsReady() bool {
	resp, err := c.APIContext.APIClient.R().Get("/")
	if err != nil {
		return false
	}

	return resp.IsSuccess() || resp.StatusCode() == http.StatusNotFound
}

func SetupTestContext(listenAddress, databaseURI string) (*TestContext, error) {
	accrualServer := NewTestAccrualServer()

	cfgInfra := app.ConfigBootstrap{
		AccrualSystemURL:   accrualServer.URL(),
		AuthTokenSecretKey: app.MustGenerateRandomSecretKey(),
		DatabaseURI:        databaseURI,
	}

	cfgRuntime := app.ConfigRuntime{
		AuthTokenLifetime:            5 * time.Minute,
		HTTPListenAddress:            listenAddress,
		HTTPShutdownTimeout:          500 * time.Millisecond,
		OrderProcessorWorkerPoolSize: 4,
	}

	depsInfra, err := app.BuildBootstrapDeps(cfgInfra)
	if err != nil {
		return nil, err
	}

	depsRuntime := app.BuildRuntimeDeps(depsInfra, cfgRuntime)

	serverURL := "http://" + listenAddress

	return &TestContext{
		APIContext: &APITestContext{
			APIClient:       NewAPIClient(serverURL),
			SecureAPIClient: NewSecureAPIClient(serverURL),
			accrualServer:   accrualServer,
			orderProcessor:  depsRuntime.OrderProcessor,
		},
		httpServer: depsRuntime.HTTPServer,
	}, nil
}
