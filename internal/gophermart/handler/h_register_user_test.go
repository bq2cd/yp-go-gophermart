package handler_test

import (
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/mocks"
)

var _ = Describe("RegisterUser", func() {
	var (
		tsctx            *TestContext
		mockUserService  *mocks.MockUserService
		mockTokenService *mocks.MockTokenService
		loginPassword    api.LoginPassword
	)

	BeforeEach(func() {
		tsctx = InitTestContext()

		mockUserService = mocks.NewMockUserService(tsctx.Ctrl)
		mockTokenService = mocks.NewMockTokenService(tsctx.Ctrl)

		tsctx.Handler = handler.NewHandler(mockUserService, mockTokenService)
		tsctx.SecurityHandler = handler.NewSecurityHandler(mockTokenService)

		tsctx.Request.Method = http.MethodPost
		tsctx.Request.Path = "/api/user/register"
	})

	JustBeforeEach(func() {
		tsctx.ProcessRequest()
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

			tsctx.Request.SetBodyJSON(loginPassword)

			expectedToken = "some-valid-auth-token"

			mockUserService.EXPECT().
				Register(domain.UserID(loginPassword.Login), domain.PasswordPlain(loginPassword.Password))
			mockTokenService.EXPECT().
				IssueToken(domain.UserID(loginPassword.Login)).
				Return(domain.Token(expectedToken), nil)
		})

		It("should return 200 OK and a token in Authorization header", func() {
			Expect(tsctx.GetStatusCode()).To(Equal(http.StatusOK))
			Expect(
				tsctx.GetHeaderValue(api.AuthorizationHeaderName),
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

				tsctx.Request.SetBodyJSON(loginPassword)
			})

			It("should return 400 Bad Request", func() {
				Expect(tsctx.GetStatusCode()).To(Equal(http.StatusBadRequest))
				Expect(tsctx.GetHeaderValue(api.AuthorizationHeaderName)).To(BeEmpty())
			})
		},
		Entry("password is missing", "user1", ""),
		Entry("login is missing", "", "password1"),
		Entry("both login and password are missing", "", ""),
		Entry("login is shorter than 4 characters", "u1", "password1"),
		Entry("password is shorter than 8 characters", "user1", "pass"),
		Entry("password is longer than 64 characters", "user1", strings.Repeat("pass1234", 9)),
	)

})
