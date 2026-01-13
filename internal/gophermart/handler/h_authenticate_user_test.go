package handler_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/mocks"
)

var _ = Describe("AuthenticateUser", func() {
	var (
		testCtx       *TestContext
		testMocks     *TestMocks
		loginPassword api.LoginPassword
	)

	BeforeEach(func() {
		testCtx = InitTestContext()
		testMocks = InitTestMocks(testCtx)

		testCtx.Request.Method = http.MethodPost
		testCtx.Request.Path = "/api/user/login"
	})

	JustBeforeEach(func() {
		testCtx.ProcessRequest()
	})

	getUserServiceMockCall := func(credentials api.LoginPassword) *mocks.MockUserServiceAuthenticateCall {
		return testMocks.UserService.EXPECT().
			Authenticate(domain.UserID(credentials.Login), domain.PasswordPlain(credentials.Password))
	}

	testCasesForUserRegistrationAndAuthentication(&testCtx, &testMocks, getUserServiceMockCall)

	When("user provides incorrect credentials", func() {
		BeforeEach(func() {
			loginPassword = api.LoginPassword{
				Login:    exampleUserLogin,
				Password: "incorrect-password",
			}

			testCtx.Request.SetBodyJSON(loginPassword)

			getUserServiceMockCall(loginPassword).Return(domain.ErrUserAuthenticationFailed)
		})

		expectHTTPStatusWithEmptyAuthHeader(&testCtx, http.StatusUnauthorized)
	})
})
