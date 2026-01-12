package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/mocks"
)

/////////////////////////////////////////////////////////////////////////////////

const (
	exampleUserLogin        = "user1"
	exampleUserPassword     = "password1"
	exampleValidAuthToken   = "some-valid-auth-token"
	exampleInvalidAuthToken = "some-random-string"
)

/////////////////////////////////////////////////////////////////////////////////

type TestRequest struct {
	Method string
	Path   string
	Header http.Header
	Body   io.Reader
}

func NewTestRequest() *TestRequest {
	return &TestRequest{
		Header: http.Header{},
	}
}

func (r *TestRequest) SetAuthToken(token string) {
	r.Header.Set(api.AuthorizationHeaderName, api.AuthorizationHeaderValuePrefix+token)
}

func (r *TestRequest) SetBodyJSON(obj any) {
	GinkgoHelper()

	content, err := json.Marshal(obj)
	Expect(err).NotTo(HaveOccurred())

	r.Body = bytes.NewReader(content)
}

func (r *TestRequest) SetBodyPlain(data string) {
	r.Body = strings.NewReader(data)
}

/////////////////////////////////////////////////////////////////////////////////

type TestContext struct {
	Ctrl            *gomock.Controller
	Handler         *handler.Handler
	SecurityHandler *handler.SecurityHandler
	Router          *gin.Engine
	Recorder        *httptest.ResponseRecorder
	Request         *TestRequest
}

func InitTestContext() *TestContext {
	gin.SetMode(gin.ReleaseMode)

	ctx := new(TestContext)

	ctx.Router = gin.New()
	ctx.Router.Use(gin.Recovery())

	ctx.Ctrl = gomock.NewController(GinkgoT())

	ctx.Recorder = httptest.NewRecorder()
	ctx.Request = NewTestRequest()

	return ctx
}

func (c *TestContext) ProcessRequest() {
	api.Initialize(c.Router, c.Handler, c.SecurityHandler)

	request := httptest.NewRequest(c.Request.Method, c.Request.Path, c.Request.Body)
	maps.Copy(request.Header, c.Request.Header)

	c.Router.ServeHTTP(c.Recorder, request)
}

func (c *TestContext) GetStatusCode() int {
	return c.Recorder.Result().StatusCode
}

func (c *TestContext) GetHeaderValue(name string) string {
	return c.Recorder.Result().Header.Get(name)
}

func (c *TestContext) GetBodyBytes() []byte {
	return c.Recorder.Body.Bytes()
}

/////////////////////////////////////////////////////////////////////////////////

type TestMocks struct {
	TokenService   *mocks.MockTokenService
	UserService    *mocks.MockUserService
	BalanceService *mocks.MockBalanceService
	OrderService   *mocks.MockOrderService
}

func InitTestMocks(testCtx *TestContext) *TestMocks {
	testMocks := &TestMocks{
		TokenService:   mocks.NewMockTokenService(testCtx.Ctrl),
		UserService:    mocks.NewMockUserService(testCtx.Ctrl),
		BalanceService: mocks.NewMockBalanceService(testCtx.Ctrl),
		OrderService:   mocks.NewMockOrderService(testCtx.Ctrl),
	}

	testCtx.Handler = handler.NewHandler(
		testMocks.TokenService,
		testMocks.UserService,
		testMocks.BalanceService,
		testMocks.OrderService,
	)
	testCtx.SecurityHandler = handler.NewSecurityHandler(testMocks.TokenService)

	return testMocks
}

/////////////////////////////////////////////////////////////////////////////////

func UnmashalBodyJSON[T any](data []byte) T {
	GinkgoHelper()

	out := new(T)
	err := json.Unmarshal(data, out)

	Expect(err).NotTo(HaveOccurred())

	return *out
}

/////////////////////////////////////////////////////////////////////////////////

func expectHTTPStatusWithEmptyBody(testCtxPtr **TestContext, status int) {
	var (
		testCtx *TestContext
	)

	BeforeEach(func() {
		testCtx = *testCtxPtr
	})

	It(fmt.Sprintf("should return %d %s", status, http.StatusText(status)), func() {
		Expect(testCtx.GetStatusCode()).To(Equal(status))
		Expect(testCtx.GetBodyBytes()).To(BeEmpty())
	})
}

func expectHTTPStatusWithJSONBody[T any](testCtxPtr **TestContext, status int, expected *T) {
	var (
		testCtx *TestContext
	)

	BeforeEach(func() {
		testCtx = *testCtxPtr
	})

	It(fmt.Sprintf("should return %d %s with JSON-encoded %T", status, http.StatusText(status), *expected), func() {
		Expect(testCtx.GetStatusCode()).To(Equal(status))
		Expect(UnmashalBodyJSON[T](testCtx.GetBodyBytes())).To(Equal(*expected))
	})
}

func expectHTTPStatusWithEmptyAuthHeader(testCtxPtr **TestContext, status int) {
	var (
		testCtx *TestContext
	)

	BeforeEach(func() {
		testCtx = *testCtxPtr
	})

	It(fmt.Sprintf("should return %d %s", status, http.StatusText(status)), func() {
		Expect(testCtx.GetStatusCode()).To(Equal(status))
		Expect(testCtx.GetHeaderValue(api.AuthorizationHeaderName)).To(BeEmpty())
	})
}

func expectHTTPStatusWithAuthToken(testCtxPtr **TestContext, status int, expectedTokenPtr *string) {
	var (
		testCtx *TestContext
	)

	BeforeEach(func() {
		testCtx = *testCtxPtr
	})

	It(fmt.Sprintf("should return %d %s", status, http.StatusText(status)), func() {
		Expect(testCtx.GetStatusCode()).To(Equal(status))
		Expect(
			testCtx.GetHeaderValue(api.AuthorizationHeaderName),
		).To(Equal(api.AuthorizationHeaderValuePrefix + *expectedTokenPtr))
	})
}
