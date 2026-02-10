package handler_test

import (
	"errors"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
	"github.com/bq2cd/yp-go-gophermart/internal/test/testutil"
)

type mockCallErrorReturner[T any] interface {
	Return(err error) T
}

func testCasesForUserRegistrationAndAuthentication[T mockCallErrorReturner[T]](
	testCtxPtr **TestContext,
	testMocksPtr **TestMocks,
	getUserServiceMockCall func(api.LoginPassword) T,
) {
	var (
		testCtx       *TestContext
		testMocks     *TestMocks
		loginPassword api.LoginPassword
	)

	BeforeEach(func() {
		testCtx = *testCtxPtr
		testMocks = *testMocksPtr
	})

	When("valid login and password are provided", func() {
		var (
			expectedToken string
		)

		BeforeEach(func() {
			loginPassword = api.LoginPassword{
				Login:    exampleUserLogin,
				Password: exampleUserPassword,
			}

			testCtx.Request.SetBodyJSON(loginPassword)

			expectedToken = exampleValidAuthToken

			getUserServiceMockCall(loginPassword).Return(nil)
			testMocks.TokenService.EXPECT().
				IssueToken(testutil.MockCtx(), domain.UserID(loginPassword.Login)).
				Return(domain.Token(expectedToken), nil)
		})

		expectHTTPStatusWithAuthToken(&testCtx, http.StatusOK, &expectedToken)
	})

	DescribeTableSubtree("invalid login/password combinations",
		func(login, password string) {
			BeforeEach(func() {
				loginPassword = api.LoginPassword{
					Login:    login,
					Password: password,
				}

				testCtx.Request.SetBodyJSON(loginPassword)
			})

			expectHTTPStatusWithEmptyAuthHeader(&testCtx, http.StatusBadRequest)
		},
		Entry("password is missing", "user1", ""),
		Entry("login is missing", "", "password1"),
		Entry("both login and password are missing", "", ""),
		Entry("login is shorter than 4 characters", "u1", "password1"),
		Entry("password is shorter than 8 characters", "user1", "pass"),
		Entry("password is longer than 64 characters", "user1", strings.Repeat("pass1234", 9)),
	)

	DescribeTableSubtree("internal error happens",
		func(setupMock func()) {
			BeforeEach(func() {
				loginPassword = api.LoginPassword{
					Login:    exampleUserLogin,
					Password: exampleUserPassword,
				}

				testCtx.Request.SetBodyJSON(loginPassword)

				setupMock()
			})

			expectHTTPStatusWithEmptyAuthHeader(&testCtx, http.StatusInternalServerError)
		},
		Entry("when authenticating/registering a user", func() {
			getUserServiceMockCall(loginPassword).Return(errors.New("operation failed"))
		}),
		Entry("when issuing a token", func() {
			getUserServiceMockCall(loginPassword).Return(nil)
			testMocks.TokenService.EXPECT().IssueToken(testutil.MockCtx(), domain.UserID(loginPassword.Login)).
				Return(domain.Token(""), errors.New("cannot issue token"))
		}),
	)
}
