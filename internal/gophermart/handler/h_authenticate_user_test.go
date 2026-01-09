package handler_test

import (
	"errors"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
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

	When("valid login and password are provided", func() {
		var (
			expectedToken string
		)

		BeforeEach(func() {
			loginPassword = api.LoginPassword{
				Login:    "user1",
				Password: "password1",
			}

			testCtx.Request.SetBodyJSON(loginPassword)

			expectedToken = "some-valid-auth-token"

			testMocks.UserService.EXPECT().
				Authenticate(domain.UserID(loginPassword.Login), domain.PasswordPlain(loginPassword.Password)).
				Return(nil)
			testMocks.TokenService.EXPECT().
				IssueToken(domain.UserID(loginPassword.Login)).
				Return(domain.Token(expectedToken), nil)
		})

		It("should return 200 OK and a token in Authorization header", func() {
			Expect(testCtx.GetStatusCode()).To(Equal(http.StatusOK))
			Expect(
				testCtx.GetHeaderValue(api.AuthorizationHeaderName),
			).To(Equal(api.AuthorizationHeaderValuePrefix + expectedToken))
		})
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

			It("should return 400 Bad Request", func() {
				Expect(testCtx.GetStatusCode()).To(Equal(http.StatusBadRequest))
				Expect(testCtx.GetHeaderValue(api.AuthorizationHeaderName)).To(BeEmpty())
			})
		},
		Entry("password is missing", "user1", ""),
		Entry("login is missing", "", "password1"),
		Entry("both login and password are missing", "", ""),
		Entry("login is shorter than 4 characters", "u1", "password1"),
		Entry("password is shorter than 8 characters", "user1", "pass"),
		Entry("password is longer than 64 characters", "user1", strings.Repeat("pass1234", 9)),
	)

	When("user provides incorrect credentials", func() {
		BeforeEach(func() {
			loginPassword = api.LoginPassword{
				Login:    "user2",
				Password: "incorrect-password",
			}

			testCtx.Request.SetBodyJSON(loginPassword)

			testMocks.UserService.EXPECT().
				Authenticate(domain.UserID(loginPassword.Login), domain.PasswordPlain(loginPassword.Password)).
				Return(domain.ErrUserAuthenticationFailed)
		})

		It("should return 401 Unauthorized", func() {
			Expect(testCtx.GetStatusCode()).To(Equal(http.StatusUnauthorized))
			Expect(testCtx.GetHeaderValue(api.AuthorizationHeaderName)).To(BeEmpty())
		})
	})

	DescribeTableSubtree("internal error happens",
		func(setupMock func()) {
			BeforeEach(func() {
				loginPassword = api.LoginPassword{
					Login:    "user3",
					Password: "password3",
				}

				testCtx.Request.SetBodyJSON(loginPassword)

				setupMock()
			})

			It("should return 500 Internal Server Error", func() {
				Expect(testCtx.GetStatusCode()).To(Equal(http.StatusInternalServerError))
				Expect(testCtx.GetHeaderValue(api.AuthorizationHeaderName)).To(BeEmpty())
			})
		},
		Entry("when authenticating a user", func() {
			testMocks.UserService.EXPECT().
				Authenticate(domain.UserID(loginPassword.Login), domain.PasswordPlain(loginPassword.Password)).
				Return(errors.New("authentication failed"))
		}),
		Entry("when issuing a token", func() {
			testMocks.UserService.EXPECT().
				Authenticate(domain.UserID(loginPassword.Login), domain.PasswordPlain(loginPassword.Password)).
				Return(nil)
			testMocks.TokenService.EXPECT().IssueToken(domain.UserID(loginPassword.Login)).
				Return(domain.Token(""), errors.New("cannot issue token"))
		}),
	)
})
