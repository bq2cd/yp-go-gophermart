package app

import (
	"github.com/gin-gonic/gin"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/workers"
)

// RuntimeDeps describes high-level dependencies needed
// for the [App] to run.
type RuntimeDeps struct {
	HTTPServer     *HTTPServer
	OrderProcessor *OrderProcessor
}

// BuildRuntimeDeps builds final [RuntimeDeps] dependencies from [BootstrapDeps] dependencies using [ConfigRuntime].
// It constructs all necessary intermediate dependencies and wires
// them in proper order to obtain final runtime dependencies.
// The runtime dependencies implement [Thread] interface
// and are run by an [App] instance.
func BuildRuntimeDeps(bootstrapDeps BootstrapDeps, config ConfigRuntime) RuntimeDeps {
	builder := &runtimeBuilder{
		infra:  bootstrapDeps,
		config: config,
	}

	return builder.Build()
}

// Threads returns dependencies as an array of [Thread] interfaces.
func (r *RuntimeDeps) Threads() []Thread {
	return []Thread{
		r.HTTPServer,
		r.OrderProcessor,
	}
}

type runtimeBuilder struct {
	infra  BootstrapDeps
	config ConfigRuntime
}

// Build builds final dependencies and populates [Runtime].
func (b *runtimeBuilder) Build() RuntimeDeps {
	orderService, orderProcessor := b.buildOrderServiceAndProcessor()

	httpServer := b.buildHTTPServer(orderService)

	return RuntimeDeps{
		HTTPServer:     httpServer,
		OrderProcessor: orderProcessor,
	}
}

func (b *runtimeBuilder) buildOrderServiceAndProcessor() (*service.OrderManager, *OrderProcessor) {
	queue := workers.NewOrderQueue()
	processor := workers.NewOrderProcessor(b.infra.Storage, queue, b.infra.AccrualClient)

	orderService := service.NewOrderManager(b.infra.Storage, processor)
	orderProcessor := &OrderProcessor{
		processor: processor,
	}

	return orderService, orderProcessor
}

func (b *runtimeBuilder) buildHTTPServer(orderService handler.OrderService) *HTTPServer {
	router := b.buildRouter()

	b.buildAPI(router, orderService)

	return NewHTTPServer(
		b.config.HTTPListenAddress,
		router.Handler(),
		WithHTTPServerShutdownTimeout(b.config.HTTPShutdownTimeout),
	)
}

func (b *runtimeBuilder) buildRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	router.Use(gin.Recovery())

	return router
}

func (b *runtimeBuilder) buildAPI(router gin.IRouter, orderService handler.OrderService) {
	tokenService := service.NewTokenManager(b.infra.SecretKeyProvider, b.config.AuthTokenLifetime)

	handler := b.buildHandler(tokenService, orderService)
	securityHandler := b.buildSecurityHandler(tokenService)

	api.Initialize(router, handler, securityHandler)
}

func (b *runtimeBuilder) buildHandler(
	tokenService handler.TokenService,
	orderService handler.OrderService,
) *handler.Handler {
	userService := service.NewUserManager(b.infra.Storage)
	balanceService := service.NewBalanceManager(b.infra.Storage)

	return handler.NewHandler(tokenService, userService, balanceService, orderService)
}

func (b *runtimeBuilder) buildSecurityHandler(tokenService handler.TokenService) *handler.SecurityHandler {
	return handler.NewSecurityHandler(tokenService)
}
