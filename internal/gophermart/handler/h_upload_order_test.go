package handler_test

import (
	"errors"
	"net/http"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/mocks"
)

var _ = Describe("UploadOrder", func() {
	var (
		testCtx   *TestContext
		testMocks *TestMocks
		userLogin string
	)

	BeforeEach(func() {
		testCtx = InitTestContext()
		testMocks = InitTestMocks(testCtx)

		testCtx.Request.Method = http.MethodPost
		testCtx.Request.Path = "/api/user/orders"
	})

	JustBeforeEach(func() {
		testCtx.ProcessRequest()
	})

	Context("user is authenticated", func() {
		var (
			authToken           string
			orderID             int
			mockCreateOrderCall *mocks.MockOrderServiceCreateOrderCall
		)

		BeforeEach(func() {
			userLogin = exampleUserLogin
			authToken = exampleValidAuthToken

			testCtx.Request.SetAuthToken(authToken)

			testMocks.TokenService.EXPECT().
				ValidateToken(domain.Token(authToken)).
				Return(domain.UserID(userLogin), nil)
		})

		Context("order ID is correct", func() {
			BeforeEach(func() {
				orderID = 123456789

				testCtx.Request.SetBodyPlain(strconv.Itoa(orderID))

				mockCreateOrderCall = testMocks.OrderService.EXPECT().
					CreateOrder(domain.UserID(userLogin), domain.OrderID(orderID))
			})

			When("order ID is brand new", func() {
				BeforeEach(func() {
					mockCreateOrderCall.Return(nil)
				})

				It("should return 202 Accepted", func() {
					Expect(testCtx.GetStatusCode()).To(Equal(http.StatusAccepted))
					Expect(testCtx.GetBodyBytes()).To(BeEmpty())
				})
			})

			When("order ID has already been uploaded by the same user", func() {
				BeforeEach(func() {
					mockCreateOrderCall.Return(
						&domain.OrderIDAlreadyExistsError{
							OrderID:   domain.OrderID(orderID),
							CreatedBy: domain.UserID(userLogin),
						},
					)
				})

				It("should return 200 OK", func() {
					Expect(testCtx.GetStatusCode()).To(Equal(http.StatusOK))
					Expect(testCtx.GetBodyBytes()).To(BeEmpty())
				})
			})

			When("order ID has already been uploaded by the another user", func() {
				BeforeEach(func() {
					mockCreateOrderCall.Return(
						&domain.OrderIDAlreadyExistsError{
							OrderID:   domain.OrderID(orderID),
							CreatedBy: domain.UserID("another-user"),
						},
					)
				})

				It("should return 409 Conflict", func() {
					Expect(testCtx.GetStatusCode()).To(Equal(http.StatusConflict))
					Expect(testCtx.GetBodyBytes()).To(BeEmpty())
				})
			})

			When("order ID validation fails", func() {
				BeforeEach(func() {
					mockCreateOrderCall.Return(domain.ErrOrderIDValidationFailed)
				})

				It("should return 422 Unprocessable Entity", func() {
					Expect(testCtx.GetStatusCode()).To(Equal(http.StatusUnprocessableEntity))
					Expect(testCtx.GetBodyBytes()).To(BeEmpty())
				})
			})
		})

		DescribeTableSubtree("order ID contains errors",
			func(orderIDStr string) {
				BeforeEach(func() {
					testCtx.Request.SetBodyPlain(orderIDStr)
				})

				It("should return 400 Bad Request", func() {
					Expect(testCtx.GetStatusCode()).To(Equal(http.StatusBadRequest))
					Expect(testCtx.GetBodyBytes()).To(BeEmpty())
				})
			},
			Entry("empty order ID", ""),
			Entry("order ID is not a number", "abcde12345"),
			Entry("order ID is zero", "0"),
			Entry("order ID is negative", "-12345"),
			Entry("order ID is not integer", "123.456789"),
		)

		DescribeTableSubtree(
			"internal error happens",
			func(setupMock func()) {
				BeforeEach(func() {
					orderID = 9912345

					testCtx.Request.SetBodyPlain(strconv.Itoa(orderID))

					mockCreateOrderCall = testMocks.OrderService.EXPECT().
						CreateOrder(domain.UserID(userLogin), domain.OrderID(orderID))

					setupMock()
				})

				It("should return 500 Internal Server Error", func() {
					Expect(testCtx.GetStatusCode()).To(Equal(http.StatusInternalServerError))
					Expect(testCtx.GetBodyBytes()).To(BeEmpty())
				})
			},
			Entry("when creating new order", func() {
				mockCreateOrderCall.Return(errors.New("oops"))
			}),
		)
	})

	testCasesForUnauthorizedUser(&testCtx, &testMocks)
})
