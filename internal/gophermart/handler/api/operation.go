package api

import (
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

func (op *operationCtx) RespondBadRequest() {
	op.ginCtx.AbortWithStatus(http.StatusBadRequest)
}

func (op *operationCtx) RespondServerError() {
	op.ginCtx.AbortWithStatus(http.StatusInternalServerError)
}

func (op *operationCtx) setGinContext(c *gin.Context) {
	op.ginCtx = c
}

func (op *operationCtxSecure) RespondUnauthorized() {
	op.ginCtx.AbortWithStatus(http.StatusUnauthorized)
}

func operationGetRequest[T any](ginCtx *gin.Context) (*T, error) {
	req := new(T)

	err := ginCtx.BindJSON(req)
	if err != nil {
		return nil, fmt.Errorf("request validation error: %w", err)
	}

	return req, nil
}
