package handler_test

import (
	"errors"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
)

var _ = Describe("ListOrders", func() {
	var (
		testCtx   *TestContext
		testMocks *TestMocks
		userLogin string
	)

	BeforeEach(func() {
		testCtx = InitTestContext()
		testMocks = InitTestMocks(testCtx)

		testCtx.Request.Method = http.MethodGet
		testCtx.Request.Path = "/api/user/orders"
	})

	JustBeforeEach(func() {
		testCtx.ProcessRequest()
	})

	withUserAuthenticatedContext(&testCtx, &testMocks, &userLogin, func() {
		When("user has previously uploaded some orders", func() {
			var (
				now            time.Time
				expectedOrders []api.Order
			)

			BeforeEach(func() {
				now = time.Now()
				expectedOrders = []api.Order{
					{
						Number:     "12345",
						UploadedAt: now.UTC(),
						Status:     api.OrderStatusNew,
					},
					{
						Number:     "67890",
						UploadedAt: now.Add(-1 * time.Hour).UTC(),
						Status:     api.OrderStatusProcessing,
					},
					{
						Number:     "2039912",
						UploadedAt: now.Add(-6 * time.Hour).UTC(),
						Status:     api.OrderStatusProcessed,
						Accrual:    99.9,
					},
					{
						Number:     "111",
						UploadedAt: now.Add(-12 * time.Hour).UTC(),
						Status:     api.OrderStatusInvalid,
					},
				}

				testMocks.OrderService.EXPECT().
					GetOrders(mockCtx(), domain.UserID(userLogin)).
					Return([]domain.Order{
						{
							ID:        12345,
							Status:    domain.OrderStatusNew,
							CreatedAt: now.UTC(),
						},
						{
							ID:        67890,
							Status:    domain.OrderStatusProcessing,
							CreatedAt: now.Add(-1 * time.Hour).UTC(),
						},
						{
							ID:        2039912,
							Status:    domain.OrderStatusProcessed,
							CreatedAt: now.Add(-6 * time.Hour).UTC(),
						},
						{
							ID:        111,
							Status:    domain.OrderStatusInvalid,
							CreatedAt: now.Add(-12 * time.Hour).UTC(),
						},
					}, nil)

				testMocks.OrderService.EXPECT().
					GetAccruals(mockCtx(), domain.UserID(userLogin), []domain.OrderID{12345, 67890, 2039912, 111}).
					Return(map[domain.OrderID]float64{2039912: 99.9}, nil)
			})

			expectHTTPStatusWithJSONBody(&testCtx, http.StatusOK, &expectedOrders)
		})

		When("user has never uploaded any orders", func() {
			BeforeEach(func() {
				testMocks.OrderService.EXPECT().
					GetOrders(mockCtx(), domain.UserID(userLogin)).
					Return(nil, nil)
			})

			expectHTTPStatusWithEmptyBody(&testCtx, http.StatusNoContent)
		})

		DescribeTableSubtree("internal error happens",
			func(setupMocks func()) {
				BeforeEach(func() {
					setupMocks()
				})

				expectHTTPStatusWithEmptyBody(&testCtx, http.StatusInternalServerError)
			},
			Entry("when getting orders", func() {
				testMocks.OrderService.EXPECT().
					GetOrders(mockCtx(), domain.UserID(userLogin)).
					Return(nil, errors.New("error getting orders"))
			}),
			Entry("when getting accruals", func() {
				testMocks.OrderService.EXPECT().
					GetOrders(mockCtx(), domain.UserID(userLogin)).
					Return([]domain.Order{
						{
							ID:        123,
							Status:    domain.OrderStatusProcessed,
							CreatedAt: time.Now(),
						},
						{
							ID:        456,
							Status:    domain.OrderStatusProcessed,
							CreatedAt: time.Now().Add(-1 * time.Hour),
						},
					}, nil)
				testMocks.OrderService.EXPECT().
					GetAccruals(mockCtx(), domain.UserID(userLogin), []domain.OrderID{123, 456}).
					Return(nil, errors.New("error getting accruals"))
			}),
		)
	})

	testCasesForUnauthorizedUser(&testCtx, &testMocks)
})
