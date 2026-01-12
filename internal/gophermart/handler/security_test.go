package handler_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
)

func testCasesForUnauthorizedUser(testCtxPtr **TestContext, testMocksPtr **TestMocks) {
	var (
		testCtx   *TestContext
		testMocks *TestMocks
		userLogin string
	)

	BeforeEach(func() {
		testCtx = *testCtxPtr
		testMocks = *testMocksPtr
	})

	DescribeTableSubtree("user provides incorrect token",
		func(authHeaderValue string, setupMocks func()) {
			BeforeEach(func() {
				userLogin = exampleUserLogin

				testCtx.Request.Header.Set(api.AuthorizationHeaderName, authHeaderValue)

				setupMocks()
			})

			It("should return 401 Unauthorized", func() {
				Expect(testCtx.GetStatusCode()).To(Equal(http.StatusUnauthorized))
				Expect(testCtx.GetBodyBytes()).To(BeEmpty())
			})
		},
		Entry("random token", api.AuthorizationHeaderValuePrefix+exampleInvalidAuthToken, func() {
			testMocks.TokenService.EXPECT().
				ValidateToken(domain.Token(exampleInvalidAuthToken)).
				Return(domain.UserID(userLogin), domain.ErrUserAuthenticationFailed)
		}),
		Entry("correct token without Bearer prefix", exampleValidAuthToken, func() {}),
	)
}

func withUserAuthenticatedContext(
	testCtxPtr **TestContext,
	testMocksPtr **TestMocks,
	userLoginPtr *string,
	contextBody func(),
) {
	var (
		testCtx   *TestContext
		testMocks *TestMocks
	)

	BeforeEach(func() {
		testCtx = *testCtxPtr
		testMocks = *testMocksPtr
	})

	Context("user is authenticated", func() {
		var (
			authToken string
		)

		BeforeEach(func() {
			*userLoginPtr = exampleUserLogin
			authToken = exampleValidAuthToken

			testCtx.Request.SetAuthToken(authToken)

			testMocks.TokenService.EXPECT().
				ValidateToken(domain.Token(authToken)).
				Return(domain.UserID(*userLoginPtr), nil)
		})

		contextBody()
	})
}
