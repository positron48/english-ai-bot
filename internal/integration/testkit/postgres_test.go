package testkit

import (
	"context"
	"errors"
	"testing"
)

func TestIsDockerUnavailable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"generic", errors.New("something else"), false},
		{"cannot connect", errors.New("cannot connect to the docker daemon"), true},
		{"could not connect", errors.New("could not connect to docker"), true},
		{"docker daemon", errors.New("error: docker daemon not running"), true},
		{"connection refused", errors.New("connection refused"), true},
		{"no such host", errors.New("no such host"), true},
		{"mixed case", errors.New("Cannot Connect to the Docker Daemon"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDockerUnavailable(tt.err)
			if got != tt.want {
				t.Errorf("isDockerUnavailable(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestPostgresContainer_DSN(t *testing.T) {
	pc := &PostgresContainer{dsn: "postgres://user:pass@localhost:5432/test"}
	if got := pc.DSN(); got != "postgres://user:pass@localhost:5432/test" {
		t.Errorf("DSN() = %q, want postgres://user:pass@localhost:5432/test", got)
	}
}

func TestPostgresContainer_Terminate_NilContainer(t *testing.T) {
	pc := &PostgresContainer{container: nil}
	ctx := context.Background()
	err := pc.Terminate(ctx)
	if err != nil {
		t.Errorf("Terminate(nil container) = %v, want nil", err)
	}
}
