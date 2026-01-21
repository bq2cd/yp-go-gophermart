package integration_test

import (
	"maps"
	"net/http"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
	"github.com/bq2cd/yp-go-gophermart/internal/test/integration"
	"github.com/bq2cd/yp-go-gophermart/internal/test/testutil"
)

func describeOrderedAPIUserSpec(setupAPIContextFn func() *integration.APITestContext, stageData APIStageDataUser) {
	Describe("User service", Ordered, func() {
		var (
			apiCtx *integration.APITestContext
		)

		BeforeAll(func() {
			apiCtx = setupAPIContextFn()
		})

		var (
			login, password string
			resp            *integration.APIResponse
			err             error
			operationFn     func(string, string) (*integration.APIResponse, error)
		)

		expectSuccess := func() {
			It("should return JWT token in Authorization header", func() {
				expectAPISuccessWithAuthorizationHeader(resp, err)
			})
		}

		JustBeforeEach(func() {
			resp, err = operationFn(login, password)
		})

		DescribeTableSubtree("each user",
			func(_login, _password string) {
				BeforeEach(func() {
					login = _login
					password = _password
				})

				Context("user registration", func() {
					BeforeEach(func() {
						operationFn = apiCtx.APIClient.R().OperationRegisterUser
					})

					When("valid login/password pair is provided", expectSuccess)

					When("duplicate login is provided", func() {
						BeforeEach(func() {
							password = "pretending-to-be-another-user-password"
						})

						It("should return 409 Conflict", func() {
							expectAPIErrorWithoutAuthorizationHeader(resp, err, http.StatusConflict)
						})
					})
				})

				Context("user authentication", func() {
					BeforeEach(func() {
						operationFn = apiCtx.APIClient.R().OperationAuthenticateUser
					})

					When("valid login/password pair is provided", expectSuccess)

					When("invalid password is provided", func() {
						BeforeEach(func() {
							password = "forgotten-password"
						})

						It("should return 401 Unauthorized", func() {
							expectAPIErrorWithoutAuthorizationHeader(resp, err, http.StatusUnauthorized)
						})
					})
				})
			},
			stageData.UserEntries(),
		)
	})
}

/////////////////////////////////////////////////////////////////////////////////

type APIStageDataUser struct {
	SeedUsers map[string]string
}

func (d APIStageDataUser) UserEntries() []TableEntry {
	entries := make([]TableEntry, 0)

	for login, password := range d.SeedUsers {
		entries = append(entries, Entry(nil, login, password))
	}

	return entries
}

func (d APIStageDataUser) LoginEntries() []TableEntry {
	entries := make([]TableEntry, 0)

	for login := range d.SeedUsers {
		entries = append(entries, Entry(nil, login))
	}

	return entries
}

func (d APIStageDataUser) GetLogins() []string {
	return slices.Collect(maps.Keys(d.SeedUsers))
}

func (d APIStageDataUser) GetRandomUserLogin() string {
	logins := d.GetLogins()

	return testutil.RandomSliceElement(logins)
}

/////////////////////////////////////////////////////////////////////////////////

func expectAPISuccessWithAuthorizationHeader(resp *integration.APIResponse, err error) {
	expectHTTPResponseWithStatus(resp, err, http.StatusOK)
	Expect(resp.Header().Get(api.AuthorizationHeaderName)).To(MatchRegexp(`^Bearer .+`))
	Expect(resp.Bytes()).To(BeEmpty())
}

func expectAPIErrorWithoutAuthorizationHeader(resp *integration.APIResponse, err error, expectedStatus int) {
	expectHTTPResponseWithStatus(resp, err, expectedStatus)
	Expect(resp.Header().Get(api.AuthorizationHeaderName)).To(BeEmpty())
	Expect(resp.Bytes()).To(BeEmpty())
}
