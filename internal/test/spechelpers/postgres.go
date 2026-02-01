package spechelpers

import (
	"context"
	"time"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	postgresDockerImage     = "postgres:18-alpine"
	postgresDatabase        = "testdb"
	postgresUsername        = "postgres"
	postgresPassword        = "postgres"
	postgresWaitOccurrences = 2
	postgresStartupTimeout  = 5 * time.Second
)

// PostgresContainer wraps [postgres.PostgresContainer] to avoid
// importing `testcontainers-go` in all specs.
// Eventually, it might add some extra logic on top of its methods.
type PostgresContainer struct {
	*postgres.PostgresContainer
}

// StartPostgresContainer creates and runs a Docker container from
// `postgres:18-alpine` image.
// It returns container instance, which can be used to obtain
// connection string via [PostgresContainer.ConnectionString] method,
// and terminate it during cleanup with [PostgresContainer.Terminate].
func StartPostgresContainer(ctx context.Context) *PostgresContainer {
	g.GinkgoHelper()

	container, err := postgres.Run(ctx, postgresDockerImage,
		postgres.WithDatabase(postgresDatabase),
		postgres.WithUsername(postgresUsername),
		postgres.WithPassword(postgresPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(postgresWaitOccurrences).
				WithStartupTimeout(postgresStartupTimeout),
		),
	)
	o.Expect(err).To(o.Succeed())

	return &PostgresContainer{container}
}
