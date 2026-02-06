package app_test

import (
	"os"

	"github.com/alecthomas/kong"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/app"
)

const (
	cliArgsDefaultAccrualSystemAddress = "127.9.9.9:9876"
	cliArgsDefaultDatabaseURI          = "sqlite:/tmp/gophermart.db"
)

var _ = Describe("Cli Parse", func() {
	var (
		appCLI     app.CLI
		kongParser *kong.Kong
		cliArgs    []string
		err        error
	)

	BeforeEach(func() {
		kongParser, err = kong.New(&appCLI, kong.Vars(app.CLIVars()))

		Expect(err).To(Succeed())
	})

	JustBeforeEach(func() {
		_, err = kongParser.Parse(cliArgs)
	})

	Context("no required CLI flags are provided via cmdline", func() {
		BeforeEach(func() {
			cliArgs = []string{}
		})

		When("required CLI flags are missing both in cmdline and in env vars", func() {
			It("should fail with an error", func() {
				Expect(err).To(MatchError(ContainSubstring("missing flags:")))
			})
		})

		When("required CLI flags are present in env vars", func() {
			BeforeEach(func() {
				setEnvironmentVariable("ACCRUAL_SYSTEM_ADDRESS", cliArgsDefaultAccrualSystemAddress)
				setEnvironmentVariable("DATABASE_URI", cliArgsDefaultDatabaseURI)
			})

			It("should succeed and have proper value parsed", func() {
				Expect(err).To(Succeed())
				Expect(appCLI.AccrualSystemAddress).To(Equal(cliArgsDefaultAccrualSystemAddress))
				Expect(appCLI.DatabaseURI).To(Equal(cliArgsDefaultDatabaseURI))
			})
		})
	})

	Context("required CLI flags are provided in cmdline", func() {
		type flagDef struct {
			name  string
			value string
		}

		emptyFlagDef := flagDef{}

		BeforeEach(func() {
			cliArgs = []string{
				"-r", cliArgsDefaultAccrualSystemAddress,
				"-d", cliArgsDefaultDatabaseURI,
			}
		})

		JustAfterEach(func() {
			if err != nil {
				Expect(appCLI.AccrualSystemAddress).To(Equal(cliArgsDefaultAccrualSystemAddress))
				Expect(appCLI.DatabaseURI).To(Equal(cliArgsDefaultDatabaseURI))
			}
		})

		DescribeTableSubtree(
			"other CLI flags",
			func(cmd flagDef, env flagDef, actualValueFn func() string) {
				var expectedValue string

				BeforeEach(func() {
					if env != emptyFlagDef {
						setEnvironmentVariable(env.name, env.value)
						expectedValue = env.value
					}

					if cmd != emptyFlagDef {
						cliArgs = append(cliArgs, cmd.name, cmd.value)
						expectedValue = cmd.value
					}
				})

				It("should succeed and have proper value", func() {
					GinkgoLogr.Info("", "expectedValue", expectedValue)
					Expect(err).To(Succeed())
					Expect(actualValueFn()).To(Equal(expectedValue))
				})
			},
			// -a flag
			Entry(
				"-a provided in cmdline",
				flagDef{"-a", "127.1.1.1:1234"},
				flagDef{},
				func() string { return appCLI.ListenAddress },
			),
			Entry(
				"-a provided in env",
				flagDef{},
				flagDef{"RUN_ADDRESS", "127.2.3.4:4567"},
				func() string { return appCLI.ListenAddress },
			),
			Entry(
				"-a provided both in cmdline and env",
				flagDef{"-a", "127.2.2.2:1234"},
				flagDef{"RUN_ADDRESS", "127.5.5.5:4567"},
				func() string { return appCLI.ListenAddress },
			),
		)
	})

})

func setEnvironmentVariable(key, value string) {
	GinkgoHelper()

	origValue, exists := os.LookupEnv(key)

	err := os.Setenv(key, value)

	Expect(err).To(Succeed())

	DeferCleanup(func() {
		if exists {
			err = os.Setenv(key, origValue)
		} else {
			err = os.Unsetenv(key)
		}

		Expect(err).To(Succeed())
	})
}
