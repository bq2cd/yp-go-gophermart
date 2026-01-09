package handler_test

import (
	"errors"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
)

var _ = Describe("GetBalance", func() {
	var (
		testCtx   *TestContext
		testMocks *TestMocks
		userLogin string
	)

	BeforeEach(func() {
		testCtx = InitTestContext()
		testMocks = InitTestMocks(testCtx)

		testCtx.Request.Method = http.MethodGet
		testCtx.Request.Path = "/api/user/balance"
	})

	JustBeforeEach(func() {
		testCtx.ProcessRequest()
	})

	When("user is authenticated", func() {
		var (
			authToken       string
			expectedBalance api.Balance
		)

		BeforeEach(func() {
			userLogin = "user1"
			authToken = "some-valid-token"
			expectedBalance = api.Balance{
				Current:   7.33,
				Withdrawn: 50.7,
			}

			testCtx.Request.SetAuthToken(authToken)

			testMocks.TokenService.EXPECT().
				ValidateToken(domain.Token(authToken)).
				Return(domain.UserID(userLogin), nil)
			testMocks.BalanceService.EXPECT().
				GetBalance(domain.UserID(userLogin)).
				Return(expectedBalance.Current, nil)
			testMocks.BalanceService.EXPECT().
				GetTotalAmountWithdrawn(domain.UserID(userLogin)).
				Return(expectedBalance.Withdrawn, nil)
		})

		It("should return 200 OK with JSON-encoded balance", func() {
			Expect(testCtx.GetStatusCode()).To(Equal(http.StatusOK))
			Expect(UnmashalBodyJSON[api.Balance](testCtx.GetBodyBytes())).To(Equal(expectedBalance))
		})
	})

	DescribeTableSubtree("user provided incorrect token",
		func(authHeaderValue string, setupMocks func()) {
			BeforeEach(func() {
				userLogin = "user2"

				testCtx.Request.Header.Set(api.AuthorizationHeaderName, authHeaderValue)

				setupMocks()
			})

			It("should return 401 Unauthorized", func() {
				Expect(testCtx.GetStatusCode()).To(Equal(http.StatusUnauthorized))
				Expect(testCtx.GetBodyBytes()).To(BeEmpty())
			})
		},
		Entry("random token", "Bearer some-random-string", func() {
			testMocks.TokenService.EXPECT().
				ValidateToken(domain.Token("some-random-string")).
				Return(domain.UserID(userLogin), domain.ErrUserAuthenticationFailed)
		}),
		Entry("correct token without Bearer prefix", "some-valid-token", func() {}),
	)

	DescribeTableSubtree("internal error happens",
		func(setupMocks func()) {
			var (
				authToken string
			)
			BeforeEach(func() {
				userLogin = "user3"
				authToken = "valid-token3"

				testCtx.Request.SetAuthToken(authToken)

				testMocks.TokenService.EXPECT().
					ValidateToken(domain.Token(authToken)).
					Return(domain.UserID(userLogin), nil)

				setupMocks()
			})

			It("should return 500 Internal Server Error", func() {
				Expect(testCtx.GetStatusCode()).To(Equal(http.StatusInternalServerError))
				Expect(testCtx.GetBodyBytes()).To(BeEmpty())
			})
		},
		Entry("when getting user's balance current value", func() {
			testMocks.BalanceService.EXPECT().
				GetBalance(domain.UserID(userLogin)).
				Return(0, errors.New("balance current value error"))
		}),
		Entry("when getting user's total withdrawn amount", func() {
			testMocks.BalanceService.EXPECT().
				GetBalance(domain.UserID(userLogin)).
				Return(3.5, nil)
			testMocks.BalanceService.EXPECT().
				GetTotalAmountWithdrawn(domain.UserID(userLogin)).
				Return(0, errors.New("withdrawn amount error"))
		}),
	)
})
