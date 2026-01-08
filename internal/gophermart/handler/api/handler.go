package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// AuthorizationHeaderName holds header name for HTTP Authorization.
	AuthorizationHeaderName = "Authorization"
	// AuthorizationHeaderValuePrefix holds prefix for bearer tokens in Authorization HTTP header.
	AuthorizationHeaderValuePrefix = "Bearer "
)

// Handler defines API operations to be implemented for the API to be functional.
type Handler interface {
	RegisterUser(op *OperationRegisterUser)
	AuthenticateUser(op *OperationAuthenticateUser)
	GetBalance(op *OperationGetBalance)
	Withdraw(op *OperationWithdraw)
	ListWithdrawals(op *OperationListWithdrawals)
	ListOrders(op *OperationListOrders)
	UploadOrder(op *OperationUploadOrder)
}

// SecurityHandler defines a method to validate JWT tokens extracted from [AuthorizationHeaderName].
// This handler is attached to a Gin router group to protect sensitive API operations.
type SecurityHandler interface {
	JWTAuth(token string) error
}

type ginContextSetter[T any] interface {
	*T
	setGinContext(c *gin.Context)
}

func getGinSecurityHandler(handler SecurityHandler) gin.HandlerFunc {
	return func(ginCtx *gin.Context) {
		header := ginCtx.GetHeader(AuthorizationHeaderName)

		if !strings.HasPrefix(header, AuthorizationHeaderValuePrefix) {
			ginCtx.AbortWithStatus(http.StatusUnauthorized)

			return
		}

		token := strings.TrimPrefix(header, AuthorizationHeaderValuePrefix)

		err := handler.JWTAuth(token)
		if err != nil {
			_ = ginCtx.AbortWithError(http.StatusUnauthorized, err)
		}
	}
}

func getGinHandler[T any, PT ginContextSetter[T]](operationFn func(*T)) gin.HandlerFunc {
	return func(c *gin.Context) {
		op := PT(new(T))
		op.setGinContext(c)
		operationFn(op)
	}
}
