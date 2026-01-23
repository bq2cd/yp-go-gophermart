package integration

import (
	"net/http/httptest"
)

type TestContext struct {
	APIContext *APITestContext
	httpServer *httptest.Server
}

func (c *TestContext) StopHTTPServer() {
	c.httpServer.Close()
}
