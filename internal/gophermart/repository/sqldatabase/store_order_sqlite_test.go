package sqldatabase_test

import (
	"maps"
	"slices"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/generated"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/models"
)

const (
	exampleOrderID uint = 12345
)

var _ = Describe("StoreOrder", func() {
	var storage *sqldatabase.Storage

	Context("sqlite database", func() {
		var query gorm.Interface[models.Order]

		BeforeEach(OncePerOrdered, func() {
			tempStorage := createTempStorage()
			DeferCleanup(tempStorage.Cleanup)

			storage = tempStorage.Storage
		})

		BeforeEach(func() {
			query = sqldatabase.Query[models.Order](storage)
		})

		Context("creating an order", Ordered, func() {
			var (
				user, user2 models.User
				orderID     domain.OrderID
				startTime   time.Time
			)

			expectCorrectOrderData := func() {
				orders, err := query.
					Where(generated.Order.ID.Eq(orderID.Uint())).
					Preload("User", nil).
					Find(GinkgoT().Context())

				Expect(err).To(Succeed())
				Expect(orders).To(HaveLen(1))
				Expect(orders[0].ID).To(Equal(orderID.Uint()))
				Expect(orders[0].Status).To(Equal(domain.OrderStatusNew.Int()))
				Expect(orders[0].CreatedAt).To(BeTemporally(">=", startTime))
				Expect(orders[0].CreatedAt).To(BeTemporally("<=", time.Now()))
				Expect(orders[0].UserID).To(Equal(user.ID))
				Expect(orders[0].User).To(Equal(user))
			}

			BeforeAll(func() {
				user = ensureUserExists(storage, exampleUserLogin)
				user2 = ensureUserExists(storage, exampleUserLogin+"-r2d2")
				startTime = time.Now()
			})

			BeforeEach(func() {
				orderID = domain.OrderID(exampleOrderID)
			})

			When("order is new", func() {
				It("should create it", func() {
					created, ownerID, err := storage.CreateOrder(
						GinkgoT().Context(),
						domain.UserID(user.Login),
						orderID,
					)
					Expect(err).To(Succeed())
					Expect(ownerID.String()).To(Equal(user.Login))
					Expect(created).To(BeTrue(), "order should be created")

					expectCorrectOrderData()
				})
			})

			Context("order already exists", func() {
				var (
					targetLogin string
					created     bool
					ownerID     domain.UserID
					err         error
				)

				JustAfterEach(func() {
					created, ownerID, err = storage.CreateOrder(
						GinkgoT().Context(),
						domain.UserID(targetLogin),
						orderID,
					)

					Expect(err).To(Succeed())
					Expect(ownerID.String()).To(Equal(user.Login))
					Expect(created).To(BeFalse(), "order should not be created")

					expectCorrectOrderData()
				})

				When("owner is the same", func() {
					It("should return false", func() {
						targetLogin = user.Login
					})
				})

				When("owner is different", func() {
					It("should return false and the correct owner", func() {
						targetLogin = user2.Login
					})
				})
			})

		})

		Context("listing orders", Ordered, func() {
			var (
				user, user2 models.User
				orders      []models.Order
			)

			BeforeAll(func() {
				user = ensureUserExists(storage, exampleUserLogin)
				user2 = ensureUserExists(storage, exampleUserLogin+"-r2d2")

				for i := range 10 {
					order := ensureOrderExists(storage, user.Login, uint(100_00+i))
					orders = append(orders, order)
					time.Sleep(10 * time.Millisecond)
				}
			})

			When("user has some orders", func() {
				It("should return them in reverse chronological order", func() {
					actualOrders, err := storage.GetOrders(GinkgoT().Context(), domain.UserID(user.Login))
					Expect(err).To(Succeed())
					Expect(actualOrders).To(HaveLen(len(orders)))

					for i := range orders {
						Expect(actualOrders[i]).To(Equal(
							sqldatabase.ConvertModelOrderToDomainOrder(orders[len(orders)-i-1]),
						))
					}
				})
			})

			DescribeTable("when a user has no orders, it should return empty array",
				func(_login string) {
					actualOrders, err := storage.GetOrders(GinkgoT().Context(), domain.UserID(_login))
					Expect(err).To(Succeed())
					Expect(actualOrders).To(BeEmpty())
				},
				Entry("existing user", user2.Login),
				Entry("non-existent user", `non-existent-user`),
			)
		})

		Context("listing accruals", Ordered, func() {
			var (
				user                             models.User
				targetLogin                      string
				orders                           []models.Order
				expectedAccruals, actualAccruals map[domain.OrderID]float64
				err                              error
			)

			BeforeAll(func() {
				user = ensureUserExists(storage, exampleUserLogin)
				targetLogin = user.Login
				expectedAccruals = make(map[domain.OrderID]float64)

				for i := range 10 {
					order := ensureOrderExists(storage, user.Login, uint(100_00+i))
					orders = append(orders, order)

					if i%3 == 0 {
						amount := float64(i) / 100.0
						ensureAccrualExists(storage, order, amount)
						expectedAccruals[domain.OrderID(order.ID)] = amount
					}
				}
			})

			JustBeforeEach(func() {
				orderIDs := make([]domain.OrderID, 0, len(orders))
				for _, order := range orders {
					orderIDs = append(orderIDs, domain.OrderID(order.ID))
				}

				actualAccruals, err = storage.GetOrderAccruals(
					GinkgoT().Context(),
					domain.UserID(targetLogin),
					orderIDs,
				)
			})

			When("user exists", func() {
				It("should return only orders with accrual points", func() {
					Expect(err).To(Succeed())
					Expect(actualAccruals).To(Equal(expectedAccruals))
				})
			})

			When("user is missing", func() {
				BeforeEach(func() {
					targetLogin = "non-existent-user"
				})

				It("should return empty result", func() {
					Expect(err).To(Succeed())
					Expect(actualAccruals).To(BeEmpty())
				})
			})
		})

		Context("getting processable orders per user", Ordered, func() {
			var (
				seedData              map[string]map[uint]domain.OrderStatus
				expectedOrdersPerUser map[domain.UserID][]domain.OrderID
			)

			BeforeAll(func() {
				seedData = map[string]map[uint]domain.OrderStatus{
					"user-1": {
						100_03: domain.OrderStatusNew,
						100_05: domain.OrderStatusProcessing,
					},
					"user-2": {
						200_05: domain.OrderStatusProcessed,
						200_07: domain.OrderStatusInvalid,
					},
					"user-3": {
						300_00: domain.OrderStatusProcessing,
						300_02: domain.OrderStatusProcessed,
						300_04: domain.OrderStatusNew,
					},
					"user-4": {},
				}
				expectedOrdersPerUser = map[domain.UserID][]domain.OrderID{
					"user-1": {100_03, 100_05},
					"user-3": {300_00, 300_04},
				}

				for login, orders := range seedData {
					user := ensureUserExists(storage, login)

					for _, orderID := range slices.Sorted(maps.Keys(orders)) {
						order := ensureOrderExists(storage, user.Login, orderID)
						ensureOrderStatus(storage, order, orders[orderID])
					}
				}
			})

			It("should return only processable orders for all users", func() {
				orderPerUser, err := storage.GetProcessableOrdersPerUser(GinkgoT().Context())
				Expect(err).To(Succeed())
				Expect(orderPerUser).To(Equal(expectedOrdersPerUser))
			})
		})

		Context("manipulating a single order", Ordered, func() {
			var (
				userID                                  domain.UserID
				orderID, initialOrderID, missingOrderID domain.OrderID
				user                                    models.User
				err                                     error
			)

			BeforeAll(func() {
				userID = domain.UserID(exampleUserLogin)
				initialOrderID = domain.OrderID(exampleOrderID)
				missingOrderID = domain.OrderID(9876)

				user = ensureUserExists(storage, userID.String())
				order := ensureOrderExists(storage, user.Login, initialOrderID.Uint())

				Expect(order.Status).To(Equal(domain.OrderStatusNew.Int()))
			})

			BeforeEach(func() {
				orderID = initialOrderID
			})

			Context("getting order status", func() {
				var status domain.OrderStatus

				JustBeforeEach(func() {
					status, err = storage.GetOrderStatus(GinkgoT().Context(), userID, orderID)
				})

				When("order exists", func() {
					It("should return order status", func() {
						Expect(err).To(Succeed())
						Expect(status).To(Equal(domain.OrderStatusNew))
					})
				})

				When("order does not exist", func() {
					BeforeEach(func() {
						orderID = missingOrderID
					})

					It("should return ErrOrderNotFound", func() {
						Expect(err).To(MatchError(domain.ErrOrderNotFound))
						Expect(status).To(Equal(domain.OrderStatusInvalid))
					})
				})
			})

			Context("setting order status", func() {
				expectOrderStatus := func(_status domain.OrderStatus) {
					order, err := query.Where(generated.Order.ID.Eq(orderID.Uint())).First(GinkgoT().Context())
					Expect(err).To(Succeed())
					Expect(order.Status).To(Equal(_status.Int()))
				}

				Context("marking order as invalid", func() {
					JustBeforeEach(func() {
						err = storage.MarkOrderInvalid(GinkgoT().Context(), userID, orderID)
					})

					When("order exists", func() {
						It("should set status to proper value", func() {
							Expect(err).To(Succeed())
							expectOrderStatus(domain.OrderStatusInvalid)
						})
					})

					When("order does not exist", func() {
						BeforeEach(func() {
							orderID = missingOrderID
						})

						It("should return ErrOrderNotFound", func() {
							Expect(err).To(MatchError(domain.ErrOrderNotFound))
						})
					})
				})

				Context("marking order as processing", func() {
					JustBeforeEach(func() {
						err = storage.MarkOrderProcessing(GinkgoT().Context(), userID, orderID)
					})

					When("order exists", func() {
						It("should set status to proper value", func() {
							Expect(err).To(Succeed())
							expectOrderStatus(domain.OrderStatusProcessing)
						})
					})

					When("order does not exist", func() {
						BeforeEach(func() {
							orderID = missingOrderID
						})

						It("should return ErrOrderNotFound", func() {
							Expect(err).To(MatchError(domain.ErrOrderNotFound))
						})
					})
				})

				Context("marking order as processed", func() {
					var (
						accrualPoints float64
					)

					expectAccrualPoints := func(_amount float64) {
						accruals, err := sqldatabase.Query[models.Accrual](storage).
							Where(generated.Accrual.OrderID.Eq(orderID.Uint())).
							Find(GinkgoT().Context())
						Expect(err).To(Succeed())
						Expect(accruals).To(HaveLen(1))
						Expect(accruals[0].OrderID).To(Equal(orderID.Uint()))
						Expect(accruals[0].Amount).To(Equal(_amount))
					}

					expectUserBalance := func(_amount float64) {
						balances, err := sqldatabase.Query[models.Balance](storage).
							Where(generated.Balance.UserID.Eq(user.ID)).
							Find(GinkgoT().Context())
						Expect(err).To(Succeed())
						Expect(balances).To(HaveLen(1))
						Expect(balances[0].UserID).To(Equal(user.ID))
						Expect(balances[0].Current).To(Equal(_amount))
						Expect(balances[0].Withdrawn).To(BeZero())
					}

					JustBeforeEach(func() {
						err = storage.MarkOrderProcessed(GinkgoT().Context(), userID, orderID, accrualPoints)
					})

					When("order exists", func() {
						DescribeTableSubtree("accrual points",
							func(_points, _balance float64) {
								BeforeEach(func() {
									accrualPoints = _points
								})

								It(
									"should set status to proper value, update accrual points and user's balance",
									func() {
										Expect(err).To(Succeed())
										expectOrderStatus(domain.OrderStatusProcessed)
										expectAccrualPoints(_points)
										expectUserBalance(_balance)
									},
								)
							},
							Entry(nil, 3.14, 3.14),
							Entry(nil, 2.36, 5.5),
							Entry(nil, 4.5, 10.0),
						)
					})

					When("order does not exist", func() {
						BeforeEach(func() {
							orderID = missingOrderID
						})

						It("should return ErrOrderNotFound", func() {
							Expect(err).To(MatchError(domain.ErrOrderNotFound))
						})
					})
				})
			})
		})
	})
})

/////////////////////////////////////////////////////////////////////////////////

func ensureOrderExists(storage *sqldatabase.Storage, login string, orderID uint) models.Order {
	GinkgoHelper()

	created, ownerID, err := storage.CreateOrder(
		GinkgoT().Context(),
		domain.UserID(login),
		domain.OrderID(orderID),
	)
	Expect(created).To(BeTrue(), "order should be created")
	Expect(ownerID.String()).To(Equal(login))
	Expect(err).To(Succeed())

	order, err := sqldatabase.Query[models.Order](storage).
		Where(generated.Order.ID.Eq(orderID)).
		First(GinkgoT().Context())
	Expect(err).To(Succeed())

	return order
}

func ensureAccrualExists(storage *sqldatabase.Storage, order models.Order, amount float64) {
	GinkgoHelper()

	err := sqldatabase.Query[models.Accrual](storage).
		Create(GinkgoT().Context(), &models.Accrual{
			OrderID: order.ID,
			Amount:  amount,
		})

	Expect(err).To(Succeed())
}

func ensureOrderStatus(storage *sqldatabase.Storage, order models.Order, status domain.OrderStatus) {
	GinkgoHelper()

	_, err := sqldatabase.Query[models.Order](storage).
		Where(generated.Order.ID.Eq(order.ID)).
		Set(generated.Order.Status.Set(status.Int())).
		Update(GinkgoT().Context())
	Expect(err).To(Succeed())

	order, err = sqldatabase.Query[models.Order](storage).
		Where(generated.Order.ID.Eq(order.ID)).
		First(GinkgoT().Context())
	Expect(err).To(Succeed())
	Expect(order.Status).To(Equal(status.Int()))
}
