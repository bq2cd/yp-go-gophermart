package handler_test

import (
	"errors"
	"net/http"
	"strconv"

	. "github.com/onsi/ginkgo/v2"

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

	withUserAuthenticatedContext(&testCtx, &testMocks, &userLogin, func() {
		var (
			orderID             int
			mockCreateOrderCall *mocks.MockOrderServiceCreateOrderCall
		)

		Context("order ID is correct", func() {
			BeforeEach(func() {
				orderID = exampleValidOrderID

				testCtx.Request.SetBodyPlain(strconv.Itoa(orderID))

				mockCreateOrderCall = testMocks.OrderService.EXPECT().
					CreateOrder(domain.UserID(userLogin), domain.OrderID(orderID))
			})

			When("order ID is brand new", func() {
				BeforeEach(func() {
					mockCreateOrderCall.Return(nil)
				})

				expectHTTPStatusWithEmptyBody(&testCtx, http.StatusAccepted)
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

				expectHTTPStatusWithEmptyBody(&testCtx, http.StatusOK)
			})

			When("order ID has already been uploaded by another user", func() {
				BeforeEach(func() {
					mockCreateOrderCall.Return(
						&domain.OrderIDAlreadyExistsError{
							OrderID:   domain.OrderID(orderID),
							CreatedBy: domain.UserID("another-user"),
						},
					)
				})

				expectHTTPStatusWithEmptyBody(&testCtx, http.StatusConflict)
			})

		})

		When("order ID fails Luhn's checksum validation", func() {
			BeforeEach(func() {
				testCtx.Request.SetBodyPlain("123")
			})

			expectHTTPStatusWithEmptyBody(&testCtx, http.StatusUnprocessableEntity)
		})

		DescribeTableSubtree("order ID contains errors",
			func(orderIDStr string) {
				BeforeEach(func() {
					testCtx.Request.SetBodyPlain(orderIDStr)
				})

				expectHTTPStatusWithEmptyBody(&testCtx, http.StatusBadRequest)
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
					orderID = exampleValidOrderID

					testCtx.Request.SetBodyPlain(strconv.Itoa(orderID))

					mockCreateOrderCall = testMocks.OrderService.EXPECT().
						CreateOrder(domain.UserID(userLogin), domain.OrderID(orderID))

					setupMock()
				})

				expectHTTPStatusWithEmptyBody(&testCtx, http.StatusInternalServerError)
			},
			Entry("when creating new order", func() {
				mockCreateOrderCall.Return(errors.New("oops"))
			}),
		)
	})

	testCasesForUnauthorizedUser(&testCtx, &testMocks)
})
