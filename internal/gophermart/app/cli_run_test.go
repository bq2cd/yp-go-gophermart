package app_test

import (
	"context"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/sync/errgroup"
	"resty.dev/v3"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/app"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/models"
	"github.com/bq2cd/yp-go-gophermart/internal/test/testutil"
)

const (
	exampleUserLogin    = "test-user"
	exampleUserPassword = "test-password"
)

var _ = Describe("Cli Run", func() {
	var (
		appCLI     app.CLI
		httpClient *resty.Client
	)

	BeforeEach(func() {
		addr, err := testutil.GetRandomListenAddress(GinkgoT().Context())
		Expect(err).To(Succeed())

		appCLI = app.CLI{
			ListenAddress: addr,
		}

		httpClient = resty.New().
			SetBaseURL("http://" + addr).
			SetDisableWarn(true)
	})

	runAppFn := func(verifyRunning func()) error {
		ctx, cancel := context.WithCancel(GinkgoT().Context())

		grp := new(errgroup.Group)

		grp.Go(func() error {
			defer GinkgoRecover()

			return appCLI.Run(ctx)
		})

		time.Sleep(50 * time.Millisecond)

		verifyRunning()

		cancel()

		return grp.Wait()
	}

	expectHTTPRequestsToSucceed := func() {
		GinkgoHelper()

		resp, err := httpClient.R().Get("/api/user/balance")

		Expect(err).To(Succeed())
		Expect(resp.StatusCode()).To(Equal(http.StatusUnauthorized))

		resp, err = httpClient.R().
			SetBody(api.LoginPassword{Login: exampleUserLogin, Password: exampleUserPassword}).
			Post("/api/user/register")

		Expect(err).To(Succeed())
		Expect(resp.StatusCode()).To(Equal(http.StatusOK))

		token := resp.Header().Get(api.AuthorizationHeaderName)

		resp, err = httpClient.R().
			SetAuthToken(strings.TrimPrefix(token, api.AuthorizationHeaderValuePrefix)).
			Get("/api/user/balance")

		Expect(err).To(Succeed())
		Expect(resp.StatusCode()).To(Equal(http.StatusOK))
	}

	Context("starting HTTP server", func() {
		var (
			err      error
			verifyFn func()
		)

		JustBeforeEach(func() {
			err = runAppFn(verifyFn)
		})

		Context("failure scenarios", func() {
			BeforeEach(func() {
				verifyFn = func() {}
			})

			Context("accrual system address missing or invalid", func() {
				When("accrual system address is missing", func() {
					It("should fail to start", func() {
						Expect(err).To(MatchError(ContainSubstring("accrual system address cannot be empty")))
					})
				})

				When("accrual system address is malformed", func() {
					BeforeEach(func() {
						appCLI.AccrualSystemAddress = ":::"
					})

					It("should fail to start", func() {
						Expect(err).To(MatchError(ContainSubstring("cannot parse accrual system address as url")))
					})
				})
			})

			Context("accrual system address is present and valid", func() {
				BeforeEach(func() {
					appCLI.AccrualSystemAddress = "localhost:123"
				})

				When("listen address is empty", func() {
					BeforeEach(func() {
						appCLI.ListenAddress = ""
					})

					It("should fail to start", func() {
						Expect(err).To(MatchError(ContainSubstring("listen address cannot be empty")))
					})
				})

				When("secret key file points to non-existent path", func() {
					BeforeEach(func() {
						appCLI.SecretKeyFile = "/non-existent-path-ever"
					})

					It("should fail to start", func() {
						Expect(err).To(MatchError(ContainSubstring("cannot read secret key file contents")))
					})
				})

				When("secret key file points to a directory", func() {
					BeforeEach(func() {
						appCLI.SecretKeyFile = "/tmp"
					})

					It("should fail to start", func() {
						Expect(err).To(MatchError(ContainSubstring("cannot read secret key file contents")))
					})
				})
			})
		})

		Context("success scenarios", func() {
			BeforeEach(func() {
				appCLI.AccrualSystemAddress = "localhost:9999"

				verifyFn = expectHTTPRequestsToSucceed
			})

			Context("in-memory storage backend", func() {
				When("all options are valid", func() {
					It("should succeed and respond to HTTP requests", func() {
						Expect(err).To(Succeed())
					})
				})
			})

			Context("sqlite storage backend", func() {
				var databaseURI string

				expectDatabaseStateToBeCorrect := func() {
					GinkgoHelper()

					storage, err := sqldatabase.NewStorage(databaseURI)
					Expect(err).To(Succeed())

					users, err := sqldatabase.Query[models.User](storage).
						Find(GinkgoT().Context())
					Expect(err).To(Succeed())
					Expect(users).To(HaveLen(1))
					Expect(users[0].Login).To(Equal(exampleUserLogin))
					Expect(
						bcrypt.CompareHashAndPassword(users[0].PasswordHash, []byte(exampleUserPassword)),
					).To(Succeed())
				}

				BeforeEach(func() {
					databaseURI = "sqlite:" + filepath.Join(GinkgoT().TempDir(), "gophermart.db")
					appCLI.DatabaseURI = databaseURI

					verifyFn = func() {
						expectHTTPRequestsToSucceed()
						expectDatabaseStateToBeCorrect()
					}
				})

				When("all options are valid", func() {
					It("should succeed and respond to HTTP requests", func() {
						Expect(err).To(Succeed())
					})
				})
			})

		})
	})

})
