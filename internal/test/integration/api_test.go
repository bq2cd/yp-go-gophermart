package integration_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	accdomain "github.com/bq2cd/yp-go-gophermart/internal/accrual/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
	"github.com/bq2cd/yp-go-gophermart/internal/test/integration"
	"github.com/bq2cd/yp-go-gophermart/internal/test/testutil"
)

func describeOrderedAPISpec(databaseURIPtr *string) {
	var (
		testCtx *integration.TestContext
		err     error
	)

	setupAPIContextFn := func() *integration.APITestContext {
		return testCtx.APIContext
	}

	stageDataUser := APIStageDataUser{
		SeedUsers: map[string]string{
			"test-user-1": "super-puper-password-1",
			"test-user-2": "super-puper-password-2",
		},
	}

	stageDataOrder := APIStageDataOrder{
		APIStageDataUser: stageDataUser,
		SeedAccrualData: integration.TestAccrualData{
			1001239: {Status: accdomain.OrderStatusRegistered},
			1003458: {Status: accdomain.OrderStatusProcessing},
			1005677: {Status: accdomain.OrderStatusProcessed, Accrual: 0.0},
			1007897: {Status: accdomain.OrderStatusProcessed, Accrual: 9.7},
			1009000: {Status: accdomain.OrderStatusInvalid},
			// 2002343: // missing in the accrual system,
			// 2004562: // missing in the accrual system,
			2006781: {Status: accdomain.OrderStatusProcessed, Accrual: 8.1},
			2008902: {Status: accdomain.OrderStatusProcessed, Accrual: 0.2},
		},
		SeedOrders: map[string][]domain.OrderID{
			"test-user-1": {
				1001239,
				1003458,
				1005677,
				1007897,
				1009000,
			},
			"test-user-2": {
				2002343,
				2004562,
				2006781,
				2008902,
			},
		},
		ExpectedAPIOrders: map[string][]api.Order{
			"test-user-1": {
				{Number: "1009000", Status: api.OrderStatusInvalid},
				{Number: "1007897", Status: api.OrderStatusProcessed, Accrual: 9.7},
				{Number: "1005677", Status: api.OrderStatusProcessed},
				{Number: "1003458", Status: api.OrderStatusProcessing},
				{Number: "1001239", Status: api.OrderStatusProcessing},
			},
			"test-user-2": {
				{Number: "2008902", Status: api.OrderStatusProcessed, Accrual: 0.2},
				{Number: "2006781", Status: api.OrderStatusProcessed, Accrual: 8.1},
				{Number: "2004562", Status: api.OrderStatusNew},
				{Number: "2002343", Status: api.OrderStatusNew},
			},
		},
		SeedAccrualData2: integration.TestAccrualData{
			1001239: {Status: accdomain.OrderStatusProcessed, Accrual: 2.39},
			1003458: {Status: accdomain.OrderStatusInvalid},
			2002343: {Status: accdomain.OrderStatusProcessed, Accrual: 3.43},
			2004562: {Status: accdomain.OrderStatusProcessed, Accrual: 5.62},
		},
		ExpectedAPIOrders2: map[string][]api.Order{
			"test-user-1": {
				{Number: "1009000", Status: api.OrderStatusInvalid},
				{Number: "1007897", Status: api.OrderStatusProcessed, Accrual: 9.7},
				{Number: "1005677", Status: api.OrderStatusProcessed},
				{Number: "1003458", Status: api.OrderStatusInvalid},
				{Number: "1001239", Status: api.OrderStatusProcessed, Accrual: 2.39},
			},
			"test-user-2": {
				{Number: "2008902", Status: api.OrderStatusProcessed, Accrual: 0.2},
				{Number: "2006781", Status: api.OrderStatusProcessed, Accrual: 8.1},
				{Number: "2004562", Status: api.OrderStatusProcessed, Accrual: 5.62},
				{Number: "2002343", Status: api.OrderStatusProcessed, Accrual: 3.43},
			},
		},
	}

	stageDataBalance := APIStageDataBalance{
		APIStageDataUser: stageDataUser,
		SeedWithdrawals: map[string][]api.WithdrawalRequest{
			"test-user-1": {
				{Order: "1101237", Sum: 1.1},
				{Order: "1104561", Sum: 2.2},
				{Order: "1107895", Sum: 3.3},
			},
			"test-user-2": {
				{Order: "2101236", Sum: 2.25},
				{Order: "2104560", Sum: 2.75},
				{Order: "2107894", Sum: 3.01},
			},
		},
		ExpectedBalancesBeforeWithdrawals: map[string]api.Balance{
			"test-user-1": {
				Current:   12.09,
				Withdrawn: 0.0,
			},
			"test-user-2": {
				Current:   17.35,
				Withdrawn: 0.0,
			},
		},
		ExpectedBalancesAfterWithdrawals: map[string]api.Balance{
			"test-user-1": {
				Current:   5.49,
				Withdrawn: 6.6,
			},
			"test-user-2": {
				Current:   9.34,
				Withdrawn: 8.01,
			},
		},
		ExpectedWithdrawalTransactions: map[string][]api.WithdrawalTransaction{
			"test-user-1": {
				{Order: "1107895", Sum: 3.3},
				{Order: "1104561", Sum: 2.2},
				{Order: "1101237", Sum: 1.1},
			},
			"test-user-2": {
				{Order: "2107894", Sum: 3.01},
				{Order: "2104560", Sum: 2.75},
				{Order: "2101236", Sum: 2.25},
			},
		},
	}

	Describe("HTTP API", Ordered, func() {
		BeforeAll(func() {
			listenAddr := getRandomListenAddress()

			testCtx, err = integration.SetupTestContext(listenAddr, *databaseURIPtr)
			Expect(err).To(Succeed())

			stopFn := testCtx.Start(GinkgoT().Context())

			DeferCleanup(func() {
				Expect(stopFn()).To(Succeed())
			})

			Eventually(testCtx.IsReady).To(BeTrue(), "http server failed to start")
		})

		describeOrderedAPIUserSpec(
			setupAPIContextFn,
			stageDataUser,
		)

		Describe("authenticated operations", func() {
			BeforeAll(func() {
				for login, password := range stageDataUser.SeedUsers {
					err := testCtx.APIContext.SecureAPIClient.EnsureUserExists(login, password)
					Expect(err).To(Succeed())
				}
			})

			describeOrderedAPIOrderSpec(
				setupAPIContextFn,
				stageDataOrder,
			)

			describeOrderedAPIBalanceSpec(
				setupAPIContextFn,
				stageDataBalance,
			)
		})
	})
}

/////////////////////////////////////////////////////////////////////////////////

func getRandomListenAddress() string {
	GinkgoHelper()

	addr, err := testutil.GetRandomListenAddress(GinkgoT().Context())
	Expect(err).To(Succeed())

	return addr
}

func expectHTTPResponseWithStatus(resp *integration.APIResponse, err error, expectedStatus int) {
	Expect(err).To(Succeed())
	Expect(resp.StatusCode()).To(Equal(expectedStatus))
}

type APISecureOperationFn func(*integration.SecureAPIRequest) (*integration.APIResponse, error)

func whenUserIsNotAuthenticated(
	setupAPIContextFn func() *integration.APITestContext,
	operationProviders ...func() (string, APISecureOperationFn),
) {
	var (
		apiCtx *integration.APITestContext
	)

	entries := make([]TableEntry, 0)
	for _, provider := range operationProviders {
		desc, operationFn := provider()
		entries = append(entries, Entry(desc, operationFn))
	}

	DescribeTableSubtree("user is not authenticated",
		func(_operationFn func(*integration.SecureAPIRequest) (*integration.APIResponse, error)) {
			BeforeEach(func() {
				apiCtx = setupAPIContextFn()
			})

			It("should return 401 Unauthorized", func() {
				req := apiCtx.SecureAPIClient.R("some-random-non-existent-user")
				resp, err := _operationFn(req)

				expectHTTPResponseWithStatus(resp, err, http.StatusUnauthorized)
				Expect(resp.Bytes()).To(BeEmpty())
			})
		},
		entries,
	)
}
