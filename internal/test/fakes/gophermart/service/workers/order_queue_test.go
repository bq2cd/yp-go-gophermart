package fakes_test

import (
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/workers"
	fakes "github.com/bq2cd/yp-go-gophermart/internal/test/fakes/gophermart/service/workers"
)

var _ = Describe("OrderQueue", func() {
	var (
		queue                      *fakes.TestOrderQueue
		poppedItems, expectedItems []workers.OrderItem
	)

	populatePoppedItems := func() {
		poppedItems = []workers.OrderItem{}

		for queue.Len() > 0 {
			poppedItems = append(poppedItems, queue.PopFront())
		}
	}

	expectElementsInProperOrder := func() {
		It("should contain elements in proper order", func() {
			Expect(queue.CopyToSlice()).To(Equal(expectedItems))

			populatePoppedItems()

			Expect(poppedItems).To(Equal(expectedItems))
		})
	}

	Context("initial queue is empty", func() {
		BeforeEach(func() {
			queue = fakes.NewTestOrderQueue()

			Expect(queue.Len()).To(BeZero())
			Expect(queue.CopyToSlice()).To(BeEmpty())
		})

		When("pushing elements", func() {
			BeforeEach(func() {
				expectedItems = []workers.OrderItem{
					{UserID: "user1", OrderID: 123},
					{UserID: "user2", OrderID: 456},
					{UserID: "user3", OrderID: 789},
				}
			})

			JustBeforeEach(func() {
				for _, item := range expectedItems {
					queue.PushBack(item)
				}
			})

			expectElementsInProperOrder()
		})

		When("popping elements", func() {
			It("should panic", func() {
				Expect(func() { queue.PopFront() }).To(Panic())
			})
		})
	})

	Context("initial queue contains some elements", func() {
		var initialItems []workers.OrderItem

		BeforeEach(func() {
			initialItems = []workers.OrderItem{
				{UserID: "user1", OrderID: 123},
				{UserID: "user2", OrderID: 456},
				{UserID: "user3", OrderID: 789},
			}

			queue = fakes.NewTestOrderQueue()

			queue.CopyFromSlice(initialItems)

			Expect(queue.Len()).To(Equal(len(initialItems)))
			Expect(queue.CopyToSlice()).To(Equal(initialItems))
		})

		When("pushing new elements", func() {
			var newItems []workers.OrderItem

			BeforeEach(func() {
				newItems = []workers.OrderItem{
					{UserID: "user10", OrderID: 1230},
					{UserID: "user20", OrderID: 4560},
					{UserID: "user30", OrderID: 7890},
				}
				expectedItems = slices.Concat(initialItems, newItems)
			})

			JustBeforeEach(func() {
				for _, item := range newItems {
					queue.PushBack(item)
				}
			})

			expectElementsInProperOrder()
		})

		When("popping elements", func() {
			It("should provide elements in proper order", func() {
				populatePoppedItems()

				Expect(poppedItems).To(Equal(initialItems))
			})
		})
	})
})
