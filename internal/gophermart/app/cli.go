package app

import (
	"context"
	"fmt"
	neturl "net/url"
	"os"
	"strings"
	"time"
)

// CLIVars defines variables to use in [CLI] field tags for CLI parser to perform interpolation.
func CLIVars() map[string]string {
	return map[string]string{
		"debug_help": `Enable debug logging.`,
		"listen_address_help": `Address and port to start HTTP server at.
Example: 'localhost:8080'`,
		"database_uri_help": `Database URI to connect to.
Example: 'postgres://user:password@localhost:5432/testdb'`,
		"accrual_system_address_help": `Address and port of the accrual system server.
Example: 'localhost:8081'`,
		"secret_key_file_help": `Path to a file with secret key for JWT tokens.
Example: './secret_key.txt'`,
	}
}

// CLI is responsible for parsing command-line flags and environment variables,
// as well as launching an [App] process.
//
//nolint:lll,tagalign
type CLI struct {
	Debug                bool   `env:"DEBUG"                  help:"${debug_help}"`
	ListenAddress        string `env:"RUN_ADDRESS"            help:"${listen_address_help}"         short:"a" default:"localhost:8080"`
	DatabaseURI          string `env:"DATABASE_URI"           help:"${database_uri_help}"           short:"d"                          required:""`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" help:"${accrual_system_address_help}" short:"r"                          required:""`
	SecretKeyFile        string `env:"SECRET_KEY_FILE"        help:"${secret_key_file_help}"                                                       type:"existingfile"`
}

// Run launches main process of the application.
func (c *CLI) Run(ctx context.Context) error {
	setupLogger(c.Debug)

	app, err := c.createApp() //nolint:contextcheck
	if err != nil {
		return err
	}

	return app.Run(ctx)
}

func (c *CLI) createApp() (*App, error) {
	infraDeps, err := c.buildInfraDeps()
	if err != nil {
		return nil, err
	}

	runtimeDeps, err := c.buildRuntimeDeps(infraDeps)
	if err != nil {
		return nil, err
	}

	app := NewApp(runtimeDeps.Threads()...)

	return app, nil
}

func (c *CLI) buildInfraDeps() (BootstrapDeps, error) {
	var deps BootstrapDeps

	cfg, err := c.getBootstrapConfig()
	if err != nil {
		return deps, err
	}

	deps, err = BuildBootstrapDeps(cfg)
	if err != nil {
		return deps, err
	}

	return deps, nil
}

func (c *CLI) buildRuntimeDeps(infraDeps BootstrapDeps) (RuntimeDeps, error) {
	var deps RuntimeDeps

	cfg, err := c.getRuntimeConfig()
	if err != nil {
		return deps, err
	}

	deps = BuildRuntimeDeps(infraDeps, cfg)

	return deps, nil
}

func (c *CLI) getBootstrapConfig() (ConfigBootstrap, error) {
	var cfg ConfigBootstrap

	accrualURL, err := c.getAccrualURL()
	if err != nil {
		return cfg, err
	}

	secretKey, err := c.getSecretKeyBytes()
	if err != nil {
		return cfg, err
	}

	databaseURI, err := c.getDatabaseURI()
	if err != nil {
		return cfg, err
	}

	cfg.AccrualSystemURL = accrualURL
	cfg.AuthTokenSecretKey = secretKey
	cfg.DatabaseURI = databaseURI

	return cfg, nil
}

func (c *CLI) getRuntimeConfig() (ConfigRuntime, error) {
	var cfg ConfigRuntime

	if c.ListenAddress == "" {
		return cfg, ErrEmptyListenAddress
	}

	cfg.AuthTokenLifetime = 1 * time.Hour
	cfg.HTTPListenAddress = c.ListenAddress
	cfg.HTTPShutdownTimeout = 1 * time.Second

	return cfg, nil
}

func (c *CLI) getAccrualURL() (string, error) {
	addr := c.AccrualSystemAddress
	if addr == "" {
		return "", ErrEmptyAccrualSystemAddress
	}

	if !strings.Contains(addr, "://") {
		addr = "http://" + addr
	}

	url, err := neturl.Parse(addr)
	if err != nil {
		return "", fmt.Errorf("cannot parse accrual system address as url: %w", err)
	}

	return url.String(), nil
}

func (c *CLI) getSecretKeyBytes() ([]byte, error) {
	if c.SecretKeyFile == "" {
		return MustGenerateRandomSecretKey(), nil
	}

	contents, err := os.ReadFile(c.SecretKeyFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read secret key file contents: %w", err)
	}

	if len(contents) == 0 {
		return MustGenerateRandomSecretKey(), nil
	}

	return contents, nil
}

func (c *CLI) getDatabaseURI() (string, error) {
	databaseURI := c.DatabaseURI
	if databaseURI == "" {
		return "", ErrEmptyDatabaseURI
	}

	url, err := neturl.Parse(databaseURI)
	if err != nil {
		return "", fmt.Errorf("cannot parse database URI as url: %w", err)
	}

	return url.String(), nil
}
