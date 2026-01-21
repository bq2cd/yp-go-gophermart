package integration

import (
	"net/http/httptest"
	"time"

	"github.com/gammazero/deque"
	"github.com/gin-gonic/gin"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/workers"
	fakes "github.com/bq2cd/yp-go-gophermart/internal/test/fakes/gophermart/service/workers"
)

func SetupTestContext(storage Storage) *TestContext {
	accrualClient := setupAccrualClient()
	orderProcessor := setupOrderProcessor(accrualClient, storage)

	httpServer := setupHTTPServer(storage, orderProcessor)

	return &TestContext{
		APIContext: &APITestContext{
			APIClient:       NewAPIClient(httpServer.URL),
			SecureAPIClient: NewSecureAPIClient(httpServer.URL),
			accrualClient:   accrualClient,
			orderProcessor:  orderProcessor,
		},
		httpServer: httpServer,
	}
}

func setupAccrualClient() *fakes.TestAccrualClient {
	return fakes.NewTestAccrualClient()
}

func setupTokenService() handler.TokenService {
	secretKeyProvider := NewTestSecretKeyProvider()

	return service.NewTokenManager(secretKeyProvider, 5*time.Minute)
}

func setupOrderProcessor(accrualClient workers.AccrualClient, storage Storage) *workers.OrderProcessor {
	orderQueue := new(deque.Deque[workers.OrderItem])

	return workers.NewOrderProcessor(storage, orderQueue, accrualClient)
}

func setupHTTPServer(
	storage Storage,
	orderProcessor service.OrderProcessor,
) *httptest.Server {
	tokenService := setupTokenService()

	router := setupRouter()
	handler := setupHandler(storage, orderProcessor, tokenService)
	securityHandler := setupSecurityHandler(tokenService)

	api.Initialize(router, handler, securityHandler)

	return httptest.NewServer(router)
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	router.Use(gin.Recovery())

	return router
}

func setupHandler(
	storage Storage,
	orderProcessor service.OrderProcessor,
	tokenService handler.TokenService,
) api.Handler {
	userService := service.NewUserManager(storage)
	balanceService := service.NewBalanceManager(storage)
	orderService := service.NewOrderManager(storage, orderProcessor)

	return handler.NewHandler(tokenService, userService, balanceService, orderService)
}

func setupSecurityHandler(tokenService handler.TokenService) api.SecurityHandler {
	return handler.NewSecurityHandler(tokenService)
}
