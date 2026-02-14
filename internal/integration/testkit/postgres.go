package testkit

import (
	"context"
	"strings"
	"testing"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// PostgresContainer wraps a testcontainers Postgres instance
type PostgresContainer struct {
	container *postgres.PostgresContainer
	dsn       string
}

// DSN returns the connection string (postgres://...)
func (p *PostgresContainer) DSN() string {
	return p.dsn
}

// Terminate stops and removes the container
func (p *PostgresContainer) Terminate(ctx context.Context) error {
	if p.container != nil {
		return p.container.Terminate(ctx)
	}
	return nil
}

// StartPostgres starts a Postgres container via Testcontainers.
// Returns DSN and cleanup function. Skips the test if Docker is unavailable.
func StartPostgres(t *testing.T) (*PostgresContainer, func()) {
	t.Helper()

	ctx := context.Background()

	ctr, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("english_test"),
		postgres.WithUsername("english"),
		postgres.WithPassword("english"),
	)
	if err != nil {
		if isDockerUnavailable(err) {
			t.Skipf("Docker unavailable for Testcontainers: %v", err)
		}
		t.Fatalf("failed to start postgres container: %v", err)
	}

	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = ctr.Terminate(ctx)
		t.Fatalf("failed to get postgres connection string: %v", err)
	}

	pc := &PostgresContainer{container: ctr, dsn: dsn}
	cleanup := func() {
		_ = pc.Terminate(ctx)
	}

	return pc, cleanup
}

func isDockerUnavailable(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	checks := []string{
		"cannot connect to the docker daemon",
		"could not connect to docker",
		"docker daemon",
		"connection refused",
		"no such host",
	}
	for _, c := range checks {
		if strings.Contains(s, c) {
			return true
		}
	}
	return false
}
