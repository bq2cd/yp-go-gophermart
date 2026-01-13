package api

import (
	"github.com/gin-gonic/gin"
)

// Initialize performs basic initialization of the API.
//  1. Creates protected Gin router group with [SecurityHandler] attached.
//  2. Registers routes corresponding to API operations from [Handler],
//     with selected routes being attached to the protected group.
//  3. Configures default Gin validator.
func Initialize(router gin.IRouter, handler Handler, securityHandler SecurityHandler) {
	registerRoutes(router, handler, securityHandler)
	setupValidator()
}
