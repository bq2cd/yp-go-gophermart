package handler_test

import (
	"errors"
	"net/http"

	. "github.com/onsi/ginkgo/v2"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/mocks"
)

var _ = Describe("Withdraw", func() {
	var (
		testCtx   *TestContext
		testMocks *TestMocks
		userLogin string
	)

	BeforeEach(func() {
		testCtx = InitTestContext()
		testMocks = InitTestMocks(testCtx)

		testCtx.Request.Method = http.MethodPost
		testCtx.Request.Path = "/api/user/balance/withdraw"
	})

	JustBeforeEach(func() {
		testCtx.ProcessRequest()
	})

	withUserAuthenticatedContext(&testCtx, &testMocks, &userLogin, func() {
		var (
			withdrawalRequest  api.WithdrawalRequest
			mockWithdrawalCall *mocks.MockBalanceServicePayForOrderFromBalanceCall
		)

		Context("withdrawal request is correct", func() {
			BeforeEach(func() {
				withdrawalRequest = api.WithdrawalRequest{
					Order: "123456789",
					Sum:   12.3,
				}

				testCtx.Request.SetBodyJSON(withdrawalRequest)

				mockWithdrawalCall = testMocks.BalanceService.EXPECT().
					PayForOrderFromBalance(domain.UserID(userLogin), domain.OrderID(123456789), 12.3)
			})

			When("user has enough funds to cover the request", func() {
				BeforeEach(func() {
					mockWithdrawalCall.Return(nil)
				})

				expectHTTPStatusWithEmptyBody(&testCtx, http.StatusOK)
			})

			When("user does not have enough funds to cover the request", func() {
				BeforeEach(func() {
					mockWithdrawalCall.Return(domain.ErrBalanceNotEnoughFunds)
				})

				expectHTTPStatusWithEmptyBody(&testCtx, http.StatusPaymentRequired)
			})
		})

		DescribeTableSubtree("withdrawal request contains errors",
			func(orderID string, amount float64) {
				BeforeEach(func() {
					withdrawalRequest = api.WithdrawalRequest{
						Order: orderID,
						Sum:   amount,
					}

					testCtx.Request.SetBodyJSON(withdrawalRequest)
				})

				expectHTTPStatusWithEmptyBody(&testCtx, http.StatusUnprocessableEntity)
			},
			Entry("empty order id", "", 12.3),
			Entry("order id is not a number", "abcd123", 12.3),
			Entry("order id is zero", "0", 12.3),
			Entry("amount is zero", "123456789", 0.0),
			Entry("amount is negative", "123456789", -3.3),
			Entry("order id is zero and amount is zero", "0", 0.0),
		)

		DescribeTableSubtree(
			"internal error happens",
			func(setupMock func()) {
				BeforeEach(func() {
					withdrawalRequest = api.WithdrawalRequest{
						Order: "123456789",
						Sum:   12.3,
					}

					testCtx.Request.SetBodyJSON(withdrawalRequest)

					mockWithdrawalCall = testMocks.BalanceService.EXPECT().
						PayForOrderFromBalance(domain.UserID(userLogin), domain.OrderID(123456789), 12.3)

					setupMock()
				})

				expectHTTPStatusWithEmptyBody(&testCtx, http.StatusInternalServerError)
			},
			Entry("when order ID fails extra validation", func() {
				mockWithdrawalCall.Return(domain.ErrOrderIDValidationFailed)
			}),
			Entry("when unexpected error happens", func() {
				mockWithdrawalCall.Return(errors.New("oops"))
			}),
		)
	})

	testCasesForUnauthorizedUser(&testCtx, &testMocks)
})
