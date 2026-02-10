package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type operationCtx struct {
	ginCtx *gin.Context
}

type operationCtxSecure struct {
	operationCtx
}

// RespondBadRequest aborts request processing and returns '400 Bad Request'.
func (op *operationCtx) RespondBadRequest() {
	op.ginCtx.AbortWithStatus(http.StatusBadRequest)
}

// RespondServerError aborts request processing and returns '500 Internal Server Error'.
func (op *operationCtx) RespondServerError() {
	op.ginCtx.AbortWithStatus(http.StatusInternalServerError)
}

// Context returns [context.Context] instance corresponding to the current HTTP request.
func (op *operationCtx) Context() context.Context {
	return op.ginCtx.Request.Context()
}

func (op *operationCtx) setGinContext(c *gin.Context) {
	op.ginCtx = c
}

// GetUserID return [UserID] for a current request and
// [true] if the user ID is not empty ([false] otherwise).
func (op *operationCtxSecure) GetUserID() (UserID, bool) {
	val, exists := op.ginCtx.Get(userIDContextKeyName)
	if !exists {
		return UserIDEmptyValue, false
	}

	userID, isUserID := val.(UserID)
	if !isUserID {
		return UserIDEmptyValue, false
	}

	return userID, !userID.IsEmpty()
}

func (op *operationCtxSecure) RespondUnauthorized() {
	op.ginCtx.AbortWithStatus(http.StatusUnauthorized)
}

func operationGetRequest[T any](ginCtx *gin.Context) (*T, error) {
	req := new(T)

	err := ginCtx.ShouldBindJSON(req)
	if err != nil {
		return nil, fmt.Errorf("request validation error: %w", err)
	}

	return req, nil
}
