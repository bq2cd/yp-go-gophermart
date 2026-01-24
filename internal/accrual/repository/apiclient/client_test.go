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
					ghttp.RespondWithJSONEncoded(http.StatusOK, expectedOrder),
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
				header      http.Header
				retryConfig apiclient.RetryConfig
			)

			BeforeEach(func() {
				expectedOrder = domain.Order{
					ID:     orderID,
					Status: domain.OrderStatusRegistered,
				}

				retryConfig = apiclient.RetryConfig{
					Count:       3,
					WaitTime:    50 * time.Millisecond,
					MaxWaitTime: 500 * time.Millisecond,
				}

				header = http.Header{}

				handlers = append(handlers,
					ghttp.RespondWith(http.StatusTooManyRequests, nil, header),
				)
			})

			When("retries are disabled", func() {
				BeforeEach(func() {
					expectedOrder = domain.Order{}

					retryConfig.Count = 0

					options = append(options,
						apiclient.WithRetryConfig(retryConfig),
					)
				})

				It("should return ErrRateLimitExceeded", func() {
					Expect(err).To(MatchError(domain.ErrRateLimitExceeded))
				})
			})

			When("server does not return Retry-After header", func() {
				BeforeEach(func() {
					options = append(options,
						apiclient.WithRetryConfig(retryConfig),
					)

					server.AppendHandlers(
						ghttp.RespondWith(http.StatusTooManyRequests, nil, header),
						ghttp.RespondWith(http.StatusTooManyRequests, nil, header),
						ghttp.RespondWithJSONEncoded(http.StatusOK, expectedOrder),
					)
				})

				It("should return correct order information", func() {
					Expect(err).To(Succeed())
				})
			})

			When("server returns Retry-After header", func() {
				BeforeEach(func() {
					header.Set("Retry-After", "1")

					server.AppendHandlers(
						ghttp.RespondWith(http.StatusTooManyRequests, nil, header),
						ghttp.RespondWithJSONEncoded(http.StatusOK, expectedOrder),
					)
				})

				It("should return correct order information", func() {
					Expect(err).To(Succeed())
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
