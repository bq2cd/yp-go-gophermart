package handler_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/mocks"
	"github.com/bq2cd/yp-go-gophermart/internal/testutil"
)

var _ = Describe("RegisterUser", func() {
	var (
		testCtx       *TestContext
		testMocks     *TestMocks
		loginPassword api.LoginPassword
	)

	BeforeEach(func() {
		testCtx = InitTestContext()
		testMocks = InitTestMocks(testCtx)

		testCtx.Request.Method = http.MethodPost
		testCtx.Request.Path = "/api/user/register"
	})

	JustBeforeEach(func() {
		testCtx.ProcessRequest()
	})

	getUserServiceMockCall := func(credentials api.LoginPassword) *mocks.MockUserServiceRegisterCall {
		return testMocks.UserService.EXPECT().
			Register(testutil.MockCtx(), domain.UserID(credentials.Login), domain.PasswordPlain(credentials.Password))
	}
	testCasesForUserRegistrationAndAuthentication(&testCtx, &testMocks, getUserServiceMockCall)

	When("user with given login already exists", func() {
		BeforeEach(func() {
			loginPassword = api.LoginPassword{
				Login:    exampleUserLogin,
				Password: exampleUserPassword,
			}

			testCtx.Request.SetBodyJSON(loginPassword)

			getUserServiceMockCall(loginPassword).Return(domain.ErrUserIDConflict)
		})

		expectHTTPStatusWithEmptyAuthHeader(&testCtx, http.StatusConflict)
	})
})
