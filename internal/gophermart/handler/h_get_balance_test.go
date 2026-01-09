package handler_test

import (
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
			userLogin       string
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

})
