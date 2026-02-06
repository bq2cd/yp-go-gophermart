package sqldatabase_test

import (
	"maps"
	"slices"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/generated"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/models"
)

var _ = Describe("StoreProcessableOrder", func() {
	var storage *sqldatabase.Storage

	Context("sqlite database", func() {
		var query gorm.Interface[models.ProcessableOrder]

		BeforeEach(OncePerOrdered, func() {
			tempStorage := createTempStorage()
			DeferCleanup(tempStorage.Cleanup)

			storage = tempStorage.Storage
		})

		BeforeEach(func() {
			query = sqldatabase.Query[models.ProcessableOrder](storage)
		})

		Context("getting next processable order", Ordered, func() {
			type seedOrder struct {
				Status       domain.OrderStatus
				ProcessAfter time.Time
				Retries      uint
			}
			var (
				seedData map[string]map[uint]seedOrder
			)

			BeforeAll(func() {
				seedData = map[string]map[uint]seedOrder{
					"user-1": {
						100_03: {
							Status:       domain.OrderStatusNew,
							ProcessAfter: time.Now().Add(-4 * time.Hour),
							Retries:      1,
						},
						100_05: {Status: domain.OrderStatusProcessing},
					},
					"user-2": {
						200_05: {Status: domain.OrderStatusProcessed},
						200_07: {Status: domain.OrderStatusInvalid},
						200_08: {
							Status:       domain.OrderStatusNew,
							ProcessAfter: time.Now().Add(-8 * time.Hour),
							Retries:      2,
						},
					},
					"user-3": {
						300_00: {
							Status:       domain.OrderStatusProcessing,
							ProcessAfter: time.Now().Add(3 * time.Hour),
							Retries:      3,
						},
						300_02: {Status: domain.OrderStatusProcessed},
						300_04: {Status: domain.OrderStatusNew},
					},
					"user-4": {
						400_06: {
							Status:       domain.OrderStatusProcessing,
							ProcessAfter: time.Now().Add(-1 * time.Hour),
							Retries:      4,
						},
					},
				}

				logins := slices.Sorted(maps.Keys(seedData))
				slices.Reverse(logins)

				for _, login := range logins {
					user := ensureUserExists(storage, login)

					orderIDs := slices.Sorted(maps.Keys(seedData[login]))
					slices.Reverse(orderIDs)

					for _, orderID := range orderIDs {
						order := ensureOrderExists(storage, user.Login, orderID)
						seed := seedData[login][orderID]

						if seed.ProcessAfter.IsZero() {
							seed.ProcessAfter = order.CreatedAt
						}

						ensureOrderStatus(storage, order, seed.Status)
						ensureProcessableOrderState(storage, order, seed.ProcessAfter, seed.Retries)

						time.Sleep(10 * time.Millisecond)
					}
				}
			})

			DescribeTableSubtree(
				"querying next order",
				func(_excludeIDs []domain.OrderID, _expectOrder domain.ProcessableOrder, _expectRetries int, _expectErr error) {
					It("should return expected values", func() {
						order, retries, err := storage.GetNextProcessableOrder(GinkgoT().Context(), _excludeIDs)

						if _expectErr == nil {
							Expect(err).To(Succeed())
						} else {
							Expect(err).To(MatchError(_expectErr))
						}

						Expect(order).To(Equal(_expectOrder))
						Expect(retries).To(Equal(uint(_expectRetries)))
					})

				},
				Entry(nil, []domain.OrderID{}, domain.ProcessableOrder{UserID: "user-2", OrderID: 200_08}, 2, nil),
				Entry(nil, []domain.OrderID{}, domain.ProcessableOrder{UserID: "user-2", OrderID: 200_08}, 2, nil),
				Entry(nil, []domain.OrderID{}, domain.ProcessableOrder{UserID: "user-2", OrderID: 200_08}, 2, nil),
				Entry(
					nil,
					[]domain.OrderID{200_08},
					domain.ProcessableOrder{UserID: "user-1", OrderID: 100_03},
					1,
					nil,
				),
				Entry(
					nil,
					[]domain.OrderID{200_08, 100_03},
					domain.ProcessableOrder{UserID: "user-4", OrderID: 400_06},
					4,
					nil,
				),
				Entry(
					nil,
					[]domain.OrderID{200_08, 100_03, 400_06},
					domain.ProcessableOrder{UserID: "user-3", OrderID: 300_04},
					0,
					nil,
				),
				Entry(
					nil,
					[]domain.OrderID{200_08, 100_03, 400_06, 300_04},
					domain.ProcessableOrder{UserID: "user-1", OrderID: 100_05},
					0,
					nil,
				),
				Entry(
					nil,
					[]domain.OrderID{200_08, 100_03, 400_06, 300_04, 100_05},
					domain.ProcessableOrder{},
					0,
					sqldatabase.ErrNoProcessableOrders,
				),
			)
		})

		Context("postponing order processing", func() {
			var (
				processableOrder                          domain.ProcessableOrder
				initialProcessAfter, expectedProcessAfter time.Time
				initialRetries, expectedRetries           uint
				err, expectedErr                          error
			)

			BeforeEach(func() {
				processableOrder = domain.ProcessableOrder{
					UserID:  domain.UserID(exampleUserLogin),
					OrderID: domain.OrderID(exampleOrderID),
				}
			})

			JustBeforeEach(func() {
				user := ensureUserExists(storage, processableOrder.UserID.String())
				order := ensureOrderExists(storage, user.Login, processableOrder.OrderID.Uint())
				ensureProcessableOrderState(storage, order, initialProcessAfter, initialRetries)
			})

			expectProcessableOrderState := func(_processAfter time.Time) {
				err = storage.PostponeOrderProcessing(GinkgoT().Context(), processableOrder, _processAfter)

				if expectedErr == nil {
					Expect(err).To(Succeed())
				} else {
					Expect(err).To(MatchError(expectedErr))
				}

				state, err := query.
					Where(generated.ProcessableOrder.OrderID.Eq(processableOrder.OrderID.Uint())).
					First(GinkgoT().Context())
				Expect(err).To(Succeed())
				Expect(state.ProcessAfter).To(BeTemporally("==", expectedProcessAfter))
				Expect(state.Retries).To(Equal(expectedRetries))

			}

			When("initial state is zero", func() {
				BeforeEach(func() {
					initialProcessAfter = time.Time{}
					initialRetries = 0

					expectedProcessAfter = time.Now().Add(2 * time.Hour)
					expectedRetries = 1
				})

				It("should update order's next processable time and increment retries", func() {
					expectProcessableOrderState(expectedProcessAfter)
				})
			})

			When("new processing time is in the future", func() {
				BeforeEach(func() {
					initialProcessAfter = time.Now().Add(1 * time.Hour)
					initialRetries = 3

					expectedProcessAfter = time.Now().Add(10 * time.Hour)
					expectedRetries = 4
				})

				It("should update order's next processing time and increment retries", func() {
					expectProcessableOrderState(expectedProcessAfter)
				})
			})

			When("new processing time is in the past", func() {
				BeforeEach(func() {
					initialProcessAfter = time.Now().Add(3 * time.Hour)
					initialRetries = 5

					expectedProcessAfter = time.Now().Add(-3 * time.Hour)
					expectedRetries = 6
				})

				It("should update order's next processing time and increment retries", func() {
					expectProcessableOrderState(expectedProcessAfter)
				})
			})
		})
	})
})

/////////////////////////////////////////////////////////////////////////////////

func ensureProcessableOrderState(
	storage *sqldatabase.Storage,
	order models.Order,
	processAfter time.Time,
	retries uint,
) {
	GinkgoHelper()

	err := sqldatabase.Query[models.ProcessableOrder](storage, clause.OnConflict{
		Columns: []clause.Column{generated.ProcessableOrder.OrderID.Column()},
		DoUpdates: clause.Set{
			generated.ProcessableOrder.ProcessAfter.Set(processAfter),
			generated.ProcessableOrder.Retries.Set(retries),
		},
	}).Create(GinkgoT().Context(), &models.ProcessableOrder{
		OrderID:      order.ID,
		ProcessAfter: processAfter,
		Retries:      retries,
	})
	Expect(err).To(Succeed())
}
