package handler_test

import (
	"errors"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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

	Context("user is authenticated", func() {
		var (
			authToken          string
			withdrawalRequest  api.WithdrawalRequest
			mockWithdrawalCall *mocks.MockBalanceServicePayForOrderFromBalanceCall
		)

		BeforeEach(func() {
			userLogin = exampleUserLogin
			authToken = exampleValidAuthToken

			testCtx.Request.SetAuthToken(authToken)

			testMocks.TokenService.EXPECT().
				ValidateToken(domain.Token(authToken)).
				Return(domain.UserID(userLogin), nil)
		})

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

				It("should return 200 OK", func() {
					Expect(testCtx.GetStatusCode()).To(Equal(http.StatusOK))
					Expect(testCtx.GetBodyBytes()).To(BeEmpty())
				})
			})

			When("user does not have enough funds to cover the request", func() {
				BeforeEach(func() {
					mockWithdrawalCall.Return(domain.ErrBalanceNotEnoughFunds)
				})

				It("should return 402 Payment Required", func() {
					Expect(testCtx.GetStatusCode()).To(Equal(http.StatusPaymentRequired))
					Expect(testCtx.GetBodyBytes()).To(BeEmpty())
				})
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

				It("should return 422 Unprocessable Entity", func() {
					Expect(testCtx.GetStatusCode()).To(Equal(http.StatusUnprocessableEntity))
					Expect(testCtx.GetBodyBytes()).To(BeEmpty())
				})
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

				It("should return 500 Internal Server Error", func() {
					Expect(testCtx.GetStatusCode()).To(Equal(http.StatusInternalServerError))
					Expect(testCtx.GetBodyBytes()).To(BeEmpty())
				})
			},
			Entry("when order ID fails extra validation", func() {
				mockWithdrawalCall.Return(domain.ErrOrderIDValidationFailed)
			}),
			Entry("when unexpected error happens", func() {
				mockWithdrawalCall.Return(errors.New("oops"))
			}),
		)
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
})
