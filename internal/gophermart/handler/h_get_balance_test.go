package handler_test

import (
	"errors"
	"net/http"

	. "github.com/onsi/ginkgo/v2"

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

	withUserAuthenticatedContext(&testCtx, &testMocks, &userLogin, func() {
		When("user requests balance", func() {
			var (
				expectedBalance api.Balance
			)

			BeforeEach(func() {
				expectedBalance = api.Balance{
					Current:   7.33,
					Withdrawn: 50.7,
				}

				testMocks.BalanceService.EXPECT().
					GetBalance(domain.UserID(userLogin)).
					Return(expectedBalance.Current, nil)
				testMocks.BalanceService.EXPECT().
					GetTotalAmountWithdrawn(domain.UserID(userLogin)).
					Return(expectedBalance.Withdrawn, nil)
			})

			expectHTTPStatusWithJSONBody(&testCtx, http.StatusOK, &expectedBalance)
		})

		DescribeTableSubtree("internal error happens",
			func(setupMocks func()) {
				BeforeEach(func() {
					setupMocks()
				})

				expectHTTPStatusWithEmptyBody(&testCtx, http.StatusInternalServerError)
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

	testCasesForUnauthorizedUser(&testCtx, &testMocks)
})
