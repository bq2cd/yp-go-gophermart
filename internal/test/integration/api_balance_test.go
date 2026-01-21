package integration_test

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
	"github.com/bq2cd/yp-go-gophermart/internal/test/integration"
)

func describeOrderedAPIBalanceSpec(
	setupAPIContextFn func() *integration.APITestContext,
	stageData APIStageDataBalance,
) {
	Describe("Balance service", Ordered, func() {
		var (
			apiCtx *integration.APITestContext
		)

		BeforeAll(func() {
			apiCtx = setupAPIContextFn()

			_ = apiCtx
		})

		whenUserIsNotAuthenticated(
			setupAPIContextFn,
			func() (string, APISecureOperationFn) {
				return "getting balance value", func(_req *integration.SecureAPIRequest) (*integration.APIResponse, error) {
					return _req.OperationGetBalance()
				}
			},
			func() (string, APISecureOperationFn) {
				return "listing withdrawals", func(_req *integration.SecureAPIRequest) (*integration.APIResponse, error) {
					return _req.OperationListWithdrawals()
				}
			},
			func() (string, APISecureOperationFn) {
				return "performing a withdrawal", func(_req *integration.SecureAPIRequest) (*integration.APIResponse, error) {
					return _req.OperationWithdraw(api.WithdrawalRequest{Order: "777799990", Sum: 99.9})
				}
			},
		)

		getBalanceFn := func(_login string) api.Balance {
			resp, err := apiCtx.SecureAPIClient.R(_login).OperationGetBalance()

			expectHTTPResponseWithStatus(resp, err, http.StatusOK)

			balance, err := integration.DecodeAPIResponse[api.Balance](resp)

			Expect(err).To(Succeed())

			return balance
		}

		Context("getting current balance value before withdrawals", func() {
			DescribeTableSubtree("each user",
				func(_login string) {
					It("should return proper values", func() {
						balance := getBalanceFn(_login)

						Expect(balance).To(Equal(stageData.ExpectedBalancesBeforeWithdrawals[_login]))
					})
				},
				stageData.LoginEntries(),
			)
		})

		Context("listing withdrawal transactions before withdrawals", func() {
			DescribeTableSubtree("each user",
				func(_login string) {
					It("should return 204 No Content", func() {
						resp, err := apiCtx.SecureAPIClient.R(_login).OperationListWithdrawals()

						expectHTTPResponseWithStatus(resp, err, http.StatusNoContent)
						Expect(resp.Bytes()).To(BeEmpty())
					})
				},
				stageData.LoginEntries(),
			)
		})

		Context("performing withdrawals", func() {
			DescribeTableSubtree("each user",
				func(_login string) {
					DescribeTableSubtree("performs a withdrawal",
						func(_req api.WithdrawalRequest) {
							It("should return 200 OK", func() {
								resp, err := apiCtx.SecureAPIClient.R(_login).OperationWithdraw(_req)

								expectHTTPResponseWithStatus(resp, err, http.StatusOK)
								Expect(resp.Bytes()).To(BeEmpty())

								// Ensure a little delay to facilitate order sorting on the server side.
								time.Sleep(10 * time.Millisecond)
							})
						},
						stageData.WithdrawalRequestEntries(_login),
					)

					When("requested amount is bigger than their balance", func() {
						It("should return 402 Payment Required", func() {
							req := api.WithdrawalRequest{
								Order: "900090002", // random order ID
								Sum:   stageData.ExpectedBalancesBeforeWithdrawals[_login].Current + 999.9,
							}
							resp, err := apiCtx.SecureAPIClient.R(_login).OperationWithdraw(req)

							expectHTTPResponseWithStatus(resp, err, http.StatusPaymentRequired)
							Expect(resp.Bytes()).To(BeEmpty())
						})
					})
				},
				stageData.LoginEntries(),
			)
		})

		Context("getting current balance value after withdrawals", func() {
			DescribeTableSubtree("each user",
				func(_login string) {
					It("should return proper values", func() {
						balance := getBalanceFn(_login)

						Expect(balance).To(Equal(stageData.ExpectedBalancesAfterWithdrawals[_login]))
					})
				},
				stageData.LoginEntries(),
			)
		})

		listWithdrawalsFn := func(_login string) []api.WithdrawalTransaction {
			resp, err := apiCtx.SecureAPIClient.R(_login).OperationListWithdrawals()

			expectHTTPResponseWithStatus(resp, err, http.StatusOK)

			withdrawals, err := integration.DecodeAPIResponse[[]api.WithdrawalTransaction](resp)

			Expect(err).To(Succeed())

			return withdrawals
		}

		Context("listing withdrawal transactions after withdrawals", func() {
			DescribeTableSubtree("each user",
				func(_login string) {
					It("should return proper values", func() {
						actualTransactions := listWithdrawalsFn(_login)

						expectAPIWithdrawalsEquivalence(
							actualTransactions,
							stageData.ExpectedWithdrawalTransactions[_login],
						)
					})
				},
				stageData.LoginEntries(),
			)
		})
	})
}

/////////////////////////////////////////////////////////////////////////////////

type APIStageDataBalance struct {
	APIStageDataUser

	SeedWithdrawals                   map[string][]api.WithdrawalRequest
	ExpectedBalancesBeforeWithdrawals map[string]api.Balance
	ExpectedWithdrawalTransactions    map[string][]api.WithdrawalTransaction
	ExpectedBalancesAfterWithdrawals  map[string]api.Balance
}

func (d APIStageDataBalance) WithdrawalRequestEntries(login string) []TableEntry {
	entries := make([]TableEntry, 0)

	for _, req := range d.SeedWithdrawals[login] {
		entries = append(entries, Entry(nil, req))
	}

	return entries
}

/////////////////////////////////////////////////////////////////////////////////

type APIWithdrawalWithoutTimestamp struct {
	Order string
	Sum   float64
}

func expectAPIWithdrawalsEquivalence(actualTransactions, expectedTransactions []api.WithdrawalTransaction) {
	actualWithoutTimestamp := stripAPIWithdrawalTimestamp(actualTransactions)
	expectedWithoutTimestamp := stripAPIWithdrawalTimestamp(expectedTransactions)

	Expect(actualWithoutTimestamp).To(Equal(expectedWithoutTimestamp))
}

func stripAPIWithdrawalTimestamp(transactions []api.WithdrawalTransaction) []APIWithdrawalWithoutTimestamp {
	withoutTimestamp := []APIWithdrawalWithoutTimestamp{}
	for _, trx := range transactions {
		withoutTimestamp = append(withoutTimestamp, APIWithdrawalWithoutTimestamp{
			Order: trx.Order,
			Sum:   trx.Sum,
		})
	}

	return withoutTimestamp
}
