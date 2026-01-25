package integration_test

import (
	"context"
	"net/http"
	"slices"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
	"github.com/bq2cd/yp-go-gophermart/internal/test/integration"
	"github.com/bq2cd/yp-go-gophermart/internal/test/testutil"
)

func describeOrderedAPIOrderSpec(setupAPIContextFn func() *integration.APITestContext, stageData APIStageDataOrder) {
	Describe("Order service", Ordered, func() {
		var (
			apiCtx *integration.APITestContext
		)

		BeforeAll(func() {
			apiCtx = setupAPIContextFn()
		})

		whenUserIsNotAuthenticated(
			setupAPIContextFn,
			func() (string, APISecureOperationFn) {
				return "listing orders", func(_req *integration.SecureAPIRequest) (*integration.APIResponse, error) {
					return _req.OperationListOrders()
				}
			},
			func() (string, APISecureOperationFn) {
				return "uploading an order", func(_req *integration.SecureAPIRequest) (*integration.APIResponse, error) {
					return _req.OperationUploadOrder("888899994")
				}
			},
		)

		Context("listing orders when nothing has been uploaded yet", func() {
			DescribeTableSubtree("each user",
				func(_login string) {
					It("should return 204 No Content", func() {
						resp, err := apiCtx.SecureAPIClient.R(_login).OperationListOrders()

						expectHTTPResponseWithStatus(resp, err, http.StatusNoContent)
						Expect(resp.Bytes()).To(BeEmpty())
					})
				},
				stageData.LoginEntries(),
			)
		})

		Context("uploading new orders", func() {
			DescribeTableSubtree("each user",
				func(_login string) {
					DescribeTableSubtree("each order",
						func(_orderID domain.OrderID) {
							uploadOrderFn := func() (*integration.APIResponse, error) {
								return apiCtx.SecureAPIClient.R(_login).OperationUploadOrder(_orderID.String())
							}

							When("new order is uploaded", func() {
								It("should return 202 Accepted", func() {
									resp, err := uploadOrderFn()
									expectHTTPResponseWithStatus(resp, err, http.StatusAccepted)

									// Ensure a little delay to facilitate order sorting on the server side.
									time.Sleep(10 * time.Millisecond)
								})
							})

							When("the same order is uploaded again", func() {
								It("should return 200 OK", func() {
									resp, err := uploadOrderFn()
									expectHTTPResponseWithStatus(resp, err, http.StatusOK)
								})
							})
						},
						stageData.OrderEntries(_login),
					)
				},
				stageData.LoginEntries(),
			)

			DescribeTableSubtree("user attempts to upload an order with invalid ID",
				func(_orderID string, _status int) {
					It("should return 4xx status", func() {
						randomLogin := stageData.GetRandomUserLogin()

						resp, err := apiCtx.SecureAPIClient.R(randomLogin).OperationUploadOrder(_orderID)

						expectHTTPResponseWithStatus(resp, err, _status)
					})
				},
				Entry("empty string", "", http.StatusBadRequest),
				Entry("not a number", "abc123", http.StatusBadRequest),
				Entry("not a number, but starts with a number", "123abc", http.StatusUnprocessableEntity),
				Entry("not an integer", "3.15", http.StatusBadRequest),
				Entry("negative integer", "-22345", http.StatusBadRequest),
				Entry("zero", "0", http.StatusBadRequest),
				Entry("invalid Lunh's checksum", "123", http.StatusUnprocessableEntity),
			)

			When("user attempts to upload an order previously uploaded by another user", func() {
				It("should return 409 Conflict", func() {
					login := stageData.GetRandomUserLogin()
					otherOrders := stageData.GetOrdersForAnyUserExceptFor(login)
					order := testutil.RandomSliceElement(otherOrders)

					resp, err := apiCtx.SecureAPIClient.R(login).OperationUploadOrder(order.String())

					expectHTTPResponseWithStatus(resp, err, http.StatusConflict)
				})
			})
		})

		Context("listing uploaded orders", func() {
			var (
				serverProcessingLeeway time.Duration
			)

			listOrdersFn := func(_login string) []api.Order {
				resp, err := apiCtx.SecureAPIClient.R(_login).OperationListOrders()

				expectHTTPResponseWithStatus(resp, err, http.StatusOK)

				orders, err := integration.DecodeAPIResponse[[]api.Order](resp)

				Expect(err).To(Succeed())

				return orders
			}

			BeforeAll(func() {
				apiCtx.SetupAccrualServer(stageData.SeedAccrualData)

				serverProcessingLeeway = 100 * time.Millisecond
			})

			JustBeforeEach(func() {
				// Give the server some time to process orders.
				time.Sleep(serverProcessingLeeway)
			})

			Context("order processor is not started", func() {
				DescribeTableSubtree("each user",
					func(_login string) {
						var expectedOrders []api.Order

						BeforeEach(func() {
							expectedOrders = slices.Clone(stageData.ExpectedAPIOrders[_login])
							for i := range expectedOrders {
								expectedOrders[i].Status = api.OrderStatusNew
								expectedOrders[i].Accrual = 0
							}
						})

						It("should return all orders with NEW status", func() {
							actualOrders := listOrdersFn(_login)

							expectAPIOrdersEquivalence(actualOrders, expectedOrders)
						})
					},
					stageData.LoginEntries(),
				)
			})

			Context("order processor is working", func() {
				BeforeAll(func() {
					serverProcessingLeeway = 200 * time.Millisecond

					ctx, cancel := context.WithTimeout(GinkgoT().Context(), serverProcessingLeeway)

					DeferCleanup(cancel)

					apiCtx.StartOrderProcessing(ctx)
				})

				AfterAll(func() {
					Eventually(apiCtx.HasOrderProcessingFinished).To(BeTrue())
				})

				DescribeTableSubtree("each user",
					func(_login string) {
						It("should return all orders with proper status", func() {
							actualOrders := listOrdersFn(_login)

							expectAPIOrdersEquivalence(actualOrders, stageData.ExpectedAPIOrders[_login])
						})
					},
					stageData.LoginEntries(),
				)
			})
		})
	})
}

/////////////////////////////////////////////////////////////////////////////////

type APIStageDataOrder struct {
	APIStageDataUser

	SeedAccrualData   integration.TestAccrualData
	SeedOrders        map[string][]domain.OrderID
	ExpectedAPIOrders map[string][]api.Order
}

func (d APIStageDataOrder) GetOrdersForAnyUserExceptFor(login string) []domain.OrderID {
	logins := d.GetLogins()

	filtered := slices.DeleteFunc(logins, func(needle string) bool {
		return needle == login
	})

	otherLogin := testutil.RandomSliceElement(filtered)

	return d.SeedOrders[otherLogin]
}

func (d APIStageDataOrder) OrderEntries(login string) []TableEntry {
	entries := make([]TableEntry, 0)

	for _, orderID := range d.SeedOrders[login] {
		entries = append(entries, Entry(nil, orderID))
	}

	return entries
}

/////////////////////////////////////////////////////////////////////////////////

type APIOrderWithoutTimestamp struct {
	Accrual float64
	Number  string
	Status  api.OrderStatus
}

func expectAPIOrdersEquivalence(actualOrders, expectedOrders []api.Order) {
	actualWithoutTimestamp := stripAPIOrdersTimestamp(actualOrders)
	expectedWithoutTimestamp := stripAPIOrdersTimestamp(expectedOrders)

	Expect(actualWithoutTimestamp).To(Equal(expectedWithoutTimestamp))
}

func stripAPIOrdersTimestamp(orders []api.Order) []APIOrderWithoutTimestamp {
	withoutTimestamp := []APIOrderWithoutTimestamp{}
	for _, order := range orders {
		withoutTimestamp = append(withoutTimestamp, APIOrderWithoutTimestamp{
			Accrual: order.Accrual,
			Number:  order.Number,
			Status:  order.Status,
		})
	}

	return withoutTimestamp
}
