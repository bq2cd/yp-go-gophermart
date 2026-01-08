package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const routePrefix = "/api/user"

func registerRoutes(router gin.IRouter, handler Handler, securityHandler SecurityHandler) {
	handleRoute(router, http.MethodPost, "/register", getGinHandler(handler.RegisterUser))
	handleRoute(router, http.MethodPost, "/login", getGinHandler(handler.AuthenticateUser))

	secure := router.Group("", getGinSecurityHandler(securityHandler))

	handleRoute(secure, http.MethodGet, "/balance", getGinHandler(handler.GetBalance))
	handleRoute(secure, http.MethodPost, "/balance/withdraw", getGinHandler(handler.Withdraw))
	handleRoute(secure, http.MethodGet, "/withdrawals", getGinHandler(handler.ListWithdrawals))
	handleRoute(secure, http.MethodGet, "/orders", getGinHandler(handler.ListOrders))
	handleRoute(secure, http.MethodPost, "/orders", getGinHandler(handler.UploadOrder))
}

func handleRoute(router gin.IRouter, method, path string, handler gin.HandlerFunc) {
	router.Handle(method, routePrefix+path, handler)
}
