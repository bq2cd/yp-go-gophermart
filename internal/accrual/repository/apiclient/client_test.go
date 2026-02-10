package apiclient_test

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/bq2cd/yp-go-gophermart/internal/accrual/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/accrual/repository/apiclient"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/workers"
	"github.com/bq2cd/yp-go-gophermart/pkg/option"
)

// Ensure [apiclient.Client] implements [workers.AccrualClient] interface.
var _ workers.AccrualClient = (*apiclient.Client)(nil)

var _ = Describe("Accrual Client", func() {
	var (
		server   *ghttp.Server
		handlers []http.HandlerFunc
		client   *apiclient.Client
		options  []option.Option[apiclient.Client]
		err      error
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		handlers = []http.HandlerFunc{}
		options = []option.Option[apiclient.Client]{}
	})

	JustBeforeEach(func() {
		server.AppendHandlers(
			ghttp.CombineHandlers(handlers...),
		)

		client = apiclient.NewClient(server.URL(), options...)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("fetching order status", func() {
		var (
			orderID                    domain.OrderID
			actualOrder, expectedOrder domain.Order
		)

		BeforeEach(func() {
			orderID = 12345
			expectedOrder = domain.Order{}

			handlers = append(handlers,
				ghttp.VerifyRequest("GET", fmt.Sprintf("/api/orders/%d", orderID)),
			)
		})

		JustBeforeEach(func() {
			actualOrder, err = client.GetOrderStatus(GinkgoT().Context(), orderID)
		})

		AfterEach(func() {
			Expect(actualOrder).To(Equal(expectedOrder))
		})

		When("order exists", func() {
			BeforeEach(func() {
				expectedOrder = domain.Order{
					ID:            orderID,
					Status:        domain.OrderStatusProcessed,
					AccrualPoints: 7.77,
				}

				handlers = append(handlers,
					ghttp.RespondWithJSONEncoded(http.StatusOK, convertToOrderResponse(expectedOrder)),
				)
			})

			It("should return correct order information", func() {
				Expect(err).To(Succeed())
			})
		})

		When("order does not exist", func() {
			BeforeEach(func() {
				handlers = append(handlers,
					ghttp.RespondWith(http.StatusNoContent, nil),
				)
			})

			It("should return ErrOrderNotFound", func() {
				Expect(err).To(MatchError(domain.ErrOrderNotFound))
			})
		})

		When("accrual server hits internal error", func() {
			BeforeEach(func() {
				handlers = append(handlers,
					ghttp.RespondWith(http.StatusInternalServerError, nil),
				)
			})

			It("should return an error", func() {
				Expect(err).To(MatchError(apiclient.ErrUnexpectedHTTPStatus))
			})
		})

		Context("accrual server is busy", func() {
			var (
				header http.Header
			)

			BeforeEach(func() {
				expectedOrder = domain.Order{}

				header = http.Header{}

				handlers = append(handlers,
					ghttp.RespondWith(http.StatusTooManyRequests, nil, header),
				)
			})

			expectRetryAfterDuration := func(_duration time.Duration) {
				Expect(err).To(MatchError(
					func(_err error) bool {
						var target *domain.RateLimitExceededError

						if !errors.As(_err, &target) {
							return false
						}

						return target.RetryAfter == _duration
					},
					fmt.Sprintf("expected RateLimitExceededError with duration %v", _duration),
				))
			}

			When("server does not return Retry-After header", func() {
				BeforeEach(func() {
					expectedOrder = domain.Order{}
				})

				It("should return RateLimitExceededError with zero duration", func() {
					expectRetryAfterDuration(0)
				})
			})

			When("server returns Retry-After header", func() {
				BeforeEach(func() {
					header.Set(apiclient.RetryAfterHeaderKey, "10")

					server.AppendHandlers(
						ghttp.RespondWith(http.StatusTooManyRequests, nil, header),
					)
				})

				It("should return RateLimitExceededError with correct duration", func() {
					expectRetryAfterDuration(10 * time.Second)
				})
			})

			When("server returns incorrect Retry-After header", func() {
				BeforeEach(func() {
					header.Set(apiclient.RetryAfterHeaderKey, "-10")

					server.AppendHandlers(
						ghttp.RespondWith(http.StatusTooManyRequests, nil, header),
					)
				})

				It("should return RateLimitExceededError with default duration", func() {
					expectRetryAfterDuration(apiclient.RetryAfterDefaultDuration)
				})
			})
		})

		When("network error occurs", func() {
			BeforeEach(func() {
				faultyTransport := &faultyRoundTripper{}

				options = append(options,
					apiclient.WithTransport(faultyTransport),
				)
				handlers = append(handlers,
					ghttp.RespondWith(http.StatusNoContent, nil),
				)
			})

			It("should return an error", func() {
				Expect(err).To(MatchError(SatisfyAll(
					ContainSubstring("cannot send HTTP request"),
					ContainSubstring("some network error"),
				)))
			})
		})
	})
})

/////////////////////////////////////////////////////////////////////////////////

type faultyRoundTripper struct{}

func (rt *faultyRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return nil, errors.New("some network error")
}

func convertToOrderResponse(order domain.Order) apiclient.OrderResponse {
	GinkgoHelper()

	resp, err := apiclient.ConvertOrderToOrderResponse(order)
	Expect(err).To(Succeed())

	return resp
}
