package domain_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
)

var _ = Describe("SortByTimestamp", func() {
	var (
		now time.Time
	)

	BeforeEach(func() {
		now = time.Now().UTC()
	})

	Describe("from the newest to the oldest", func() {
		DescribeTable("sorting orders",
			func(before, after []domain.Order) {
				domain.SortByTimestampFromNewestToOldest(before)
				Expect(before).To(Equal(after))
			},
			Entry("empty slices", []domain.Order{}, []domain.Order{}),
			Entry("two orders",
				[]domain.Order{
					{ID: 123, Status: domain.OrderStatusNew, CreatedAt: now.Add(-5 * time.Hour)},
					{ID: 456, Status: domain.OrderStatusNew, CreatedAt: now.Add(-2 * time.Hour)},
				},
				[]domain.Order{
					{ID: 456, Status: domain.OrderStatusNew, CreatedAt: now.Add(-2 * time.Hour)},
					{ID: 123, Status: domain.OrderStatusNew, CreatedAt: now.Add(-5 * time.Hour)},
				},
			),
			Entry("four orders",
				[]domain.Order{
					{ID: 123, Status: domain.OrderStatusNew, CreatedAt: now.Add(-5 * time.Hour)},
					{ID: 789, Status: domain.OrderStatusNew, CreatedAt: now.Add(5 * time.Hour)},
					{ID: 456, Status: domain.OrderStatusNew, CreatedAt: now.Add(-2 * time.Hour)},
					{ID: 1111, Status: domain.OrderStatusNew, CreatedAt: now.Add(2 * time.Hour)},
				},
				[]domain.Order{
					{ID: 789, Status: domain.OrderStatusNew, CreatedAt: now.Add(5 * time.Hour)},
					{ID: 1111, Status: domain.OrderStatusNew, CreatedAt: now.Add(2 * time.Hour)},
					{ID: 456, Status: domain.OrderStatusNew, CreatedAt: now.Add(-2 * time.Hour)},
					{ID: 123, Status: domain.OrderStatusNew, CreatedAt: now.Add(-5 * time.Hour)},
				},
			),
			Entry("nine orders",
				[]domain.Order{
					{ID: 123, Status: domain.OrderStatusNew, CreatedAt: now.Add(-5 * time.Hour)},
					{ID: 789, Status: domain.OrderStatusNew, CreatedAt: now.Add(5 * time.Hour)},
					{ID: 456, Status: domain.OrderStatusNew, CreatedAt: now.Add(-2 * time.Hour)},
					{ID: 1111, Status: domain.OrderStatusNew, CreatedAt: now.Add(2 * time.Hour)},
					{ID: 1230, Status: domain.OrderStatusNew, CreatedAt: now.Add(-5 * time.Hour)},
					{ID: 7890, Status: domain.OrderStatusNew, CreatedAt: now.Add(5 * time.Hour)},
					{ID: 4560, Status: domain.OrderStatusNew, CreatedAt: now.Add(-2 * time.Hour)},
					{ID: 2222, Status: domain.OrderStatusNew, CreatedAt: now.Add(2 * time.Hour)},
				},
				[]domain.Order{
					{ID: 789, Status: domain.OrderStatusNew, CreatedAt: now.Add(5 * time.Hour)},
					{ID: 7890, Status: domain.OrderStatusNew, CreatedAt: now.Add(5 * time.Hour)},
					{ID: 1111, Status: domain.OrderStatusNew, CreatedAt: now.Add(2 * time.Hour)},
					{ID: 2222, Status: domain.OrderStatusNew, CreatedAt: now.Add(2 * time.Hour)},
					{ID: 456, Status: domain.OrderStatusNew, CreatedAt: now.Add(-2 * time.Hour)},
					{ID: 4560, Status: domain.OrderStatusNew, CreatedAt: now.Add(-2 * time.Hour)},
					{ID: 123, Status: domain.OrderStatusNew, CreatedAt: now.Add(-5 * time.Hour)},
					{ID: 1230, Status: domain.OrderStatusNew, CreatedAt: now.Add(-5 * time.Hour)},
				},
			),
		)

		DescribeTable("sorting withdrawal transactions",
			func(before, after []domain.WithdrawalTransaction) {
				domain.SortByTimestampFromNewestToOldest(before)
				Expect(before).To(Equal(after))
			},
			Entry("empty slices", []domain.WithdrawalTransaction{}, []domain.WithdrawalTransaction{}),
			Entry("two transactions",
				[]domain.WithdrawalTransaction{
					{OrderID: 123, Amount: 1.23, ProcessedAt: now.Add(-5 * time.Hour)},
					{OrderID: 456, Amount: 4.56, ProcessedAt: now.Add(-2 * time.Hour)},
				},
				[]domain.WithdrawalTransaction{
					{OrderID: 456, Amount: 4.56, ProcessedAt: now.Add(-2 * time.Hour)},
					{OrderID: 123, Amount: 1.23, ProcessedAt: now.Add(-5 * time.Hour)},
				},
			),
			Entry("seven transactions",
				[]domain.WithdrawalTransaction{
					{OrderID: 1230, Amount: 1.23, ProcessedAt: now.Add(-5 * time.Hour)},
					{OrderID: 123, Amount: 1.23, ProcessedAt: now.Add(-5 * time.Hour)},
					{OrderID: 7890, Amount: 7.89, ProcessedAt: now.Add(3 * time.Hour)},
					{OrderID: 7891, Amount: 7.89, ProcessedAt: now.Add(9 * time.Hour)},
					{OrderID: 7893, Amount: 7.89, ProcessedAt: now.Add(9 * time.Hour)},
					{OrderID: 7892, Amount: 7.89, ProcessedAt: now.Add(6 * time.Hour)},
					{OrderID: 456, Amount: 4.56, ProcessedAt: now.Add(-2 * time.Hour)},
				},
				[]domain.WithdrawalTransaction{
					{OrderID: 7891, Amount: 7.89, ProcessedAt: now.Add(9 * time.Hour)},
					{OrderID: 7893, Amount: 7.89, ProcessedAt: now.Add(9 * time.Hour)},
					{OrderID: 7892, Amount: 7.89, ProcessedAt: now.Add(6 * time.Hour)},
					{OrderID: 7890, Amount: 7.89, ProcessedAt: now.Add(3 * time.Hour)},
					{OrderID: 456, Amount: 4.56, ProcessedAt: now.Add(-2 * time.Hour)},
					{OrderID: 1230, Amount: 1.23, ProcessedAt: now.Add(-5 * time.Hour)},
					{OrderID: 123, Amount: 1.23, ProcessedAt: now.Add(-5 * time.Hour)},
				},
			),
		)
	})

	Describe("from the oldest to the newest", func() {
		DescribeTable("sorting orders",
			func(before, after []domain.Order) {
				domain.SortByTimestampFromOldestToNewest(before)
				Expect(before).To(Equal(after))
			},
			Entry("empty slices", []domain.Order{}, []domain.Order{}),
			Entry("two orders",
				[]domain.Order{
					{ID: 123, Status: domain.OrderStatusNew, CreatedAt: now.Add(-5 * time.Hour)},
					{ID: 456, Status: domain.OrderStatusNew, CreatedAt: now.Add(-2 * time.Hour)},
				},
				[]domain.Order{
					{ID: 123, Status: domain.OrderStatusNew, CreatedAt: now.Add(-5 * time.Hour)},
					{ID: 456, Status: domain.OrderStatusNew, CreatedAt: now.Add(-2 * time.Hour)},
				},
			),
			Entry("four orders",
				[]domain.Order{
					{ID: 123, Status: domain.OrderStatusNew, CreatedAt: now.Add(-5 * time.Hour)},
					{ID: 789, Status: domain.OrderStatusNew, CreatedAt: now.Add(5 * time.Hour)},
					{ID: 456, Status: domain.OrderStatusNew, CreatedAt: now.Add(-2 * time.Hour)},
					{ID: 1111, Status: domain.OrderStatusNew, CreatedAt: now.Add(2 * time.Hour)},
				},
				[]domain.Order{
					{ID: 123, Status: domain.OrderStatusNew, CreatedAt: now.Add(-5 * time.Hour)},
					{ID: 456, Status: domain.OrderStatusNew, CreatedAt: now.Add(-2 * time.Hour)},
					{ID: 1111, Status: domain.OrderStatusNew, CreatedAt: now.Add(2 * time.Hour)},
					{ID: 789, Status: domain.OrderStatusNew, CreatedAt: now.Add(5 * time.Hour)},
				},
			),
			Entry("nine orders",
				[]domain.Order{
					{ID: 123, Status: domain.OrderStatusNew, CreatedAt: now.Add(-5 * time.Hour)},
					{ID: 789, Status: domain.OrderStatusNew, CreatedAt: now.Add(5 * time.Hour)},
					{ID: 456, Status: domain.OrderStatusNew, CreatedAt: now.Add(-2 * time.Hour)},
					{ID: 1111, Status: domain.OrderStatusNew, CreatedAt: now.Add(2 * time.Hour)},
					{ID: 1230, Status: domain.OrderStatusNew, CreatedAt: now.Add(-5 * time.Hour)},
					{ID: 7890, Status: domain.OrderStatusNew, CreatedAt: now.Add(5 * time.Hour)},
					{ID: 4560, Status: domain.OrderStatusNew, CreatedAt: now.Add(-2 * time.Hour)},
					{ID: 2222, Status: domain.OrderStatusNew, CreatedAt: now.Add(2 * time.Hour)},
				},
				[]domain.Order{
					{ID: 123, Status: domain.OrderStatusNew, CreatedAt: now.Add(-5 * time.Hour)},
					{ID: 1230, Status: domain.OrderStatusNew, CreatedAt: now.Add(-5 * time.Hour)},
					{ID: 456, Status: domain.OrderStatusNew, CreatedAt: now.Add(-2 * time.Hour)},
					{ID: 4560, Status: domain.OrderStatusNew, CreatedAt: now.Add(-2 * time.Hour)},
					{ID: 1111, Status: domain.OrderStatusNew, CreatedAt: now.Add(2 * time.Hour)},
					{ID: 2222, Status: domain.OrderStatusNew, CreatedAt: now.Add(2 * time.Hour)},
					{ID: 789, Status: domain.OrderStatusNew, CreatedAt: now.Add(5 * time.Hour)},
					{ID: 7890, Status: domain.OrderStatusNew, CreatedAt: now.Add(5 * time.Hour)},
				},
			),
		)
	})
})
