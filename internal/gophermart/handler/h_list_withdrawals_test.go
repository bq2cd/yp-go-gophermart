package handler_test

import (
	"errors"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
)

var _ = Describe("ListWithdrawals", func() {
	var (
		testCtx   *TestContext
		testMocks *TestMocks
		userLogin string
	)

	BeforeEach(func() {
		testCtx = InitTestContext()
		testMocks = InitTestMocks(testCtx)

		testCtx.Request.Method = http.MethodGet
		testCtx.Request.Path = "/api/user/withdrawals"
	})

	JustBeforeEach(func() {
		testCtx.ProcessRequest()
	})

	withUserAuthenticatedContext(&testCtx, &testMocks, &userLogin, func() {
		When("user has previously made some withdrawals", func() {
			var (
				now                 time.Time
				expectedWithdrawals []api.WithdrawalTransaction
			)

			BeforeEach(func() {
				now = time.Now()
				expectedWithdrawals = []api.WithdrawalTransaction{
					{
						Order:       "12345",
						ProcessedAt: now.UTC(),
						Sum:         3.15,
					},
					{
						Order:       "67890",
						ProcessedAt: now.Add(-1 * time.Hour).UTC(),
						Sum:         7.31,
					},
				}

				testMocks.BalanceService.EXPECT().
					GetWithdrawalTransactions(mockCtx(), domain.UserID(userLogin)).
					Return([]domain.WithdrawalTransaction{
						{
							OrderID:     domain.OrderID(12345),
							Amount:      3.15,
							ProcessedAt: now.UTC(),
						},
						{
							OrderID:     domain.OrderID(67890),
							Amount:      7.31,
							ProcessedAt: now.Add(-1 * time.Hour).UTC(),
						},
					}, nil)
			})

			expectHTTPStatusWithJSONBody(&testCtx, http.StatusOK, &expectedWithdrawals)
		})

		When("user has never made any withdrawals", func() {
			BeforeEach(func() {
				testMocks.BalanceService.EXPECT().
					GetWithdrawalTransactions(mockCtx(), domain.UserID(userLogin)).
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
			Entry("when getting withdrawal transactions", func() {
				testMocks.BalanceService.EXPECT().
					GetWithdrawalTransactions(mockCtx(), domain.UserID(userLogin)).
					Return(nil, errors.New("error getting transactions"))
			}),
		)
	})

	testCasesForUnauthorizedUser(&testCtx, &testMocks)
})
