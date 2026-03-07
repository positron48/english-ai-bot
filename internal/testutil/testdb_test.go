package testutil

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestIsDockerUnavailable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"other error", errors.New("some other failure"), false},
		{"cannot connect to the docker daemon", errors.New("cannot connect to the docker daemon"), true},
		{"docker daemon mixed case", errors.New("Cannot connect to the Docker Daemon"), true},
		{"could not connect to docker", errors.New("could not connect to docker"), true},
		{"connection refused", errors.New("connection refused"), true},
		{"no such host", errors.New("dial tcp: no such host"), true},
		{"contains docker daemon", errors.New("error: Docker daemon is not running"), true},
		{"empty string error", errors.New(""), false},
		{"partial no match", errors.New("no such file"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDockerUnavailable(tt.err)
			if got != tt.want {
				t.Errorf("isDockerUnavailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTruncateAll(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	defer db.Close()

	for _, tbl := range truncateTables {
		mock.ExpectExec("TRUNCATE TABLE " + tbl + " RESTART IDENTITY CASCADE").WillReturnResult(sqlmock.NewResult(0, 0))
	}
	mock.ExpectExec(circuitBreakerInit).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(appSettingsInit).WillReturnResult(sqlmock.NewResult(0, 0))

	truncateAll(t, db)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("truncateAll() did not meet expectations: %v", err)
	}
}

// TestGetTestDSN_WhenDockerAvailable verifies GetTestDSN returns non-empty DSN.
// Skips when Docker is unavailable (e.g. CI without Docker).
func TestGetTestDSN_WhenDockerAvailable(t *testing.T) {
	dsn := GetTestDSN(t)
	if dsn == "" {
		t.Error("GetTestDSN() should return non-empty DSN when Docker is available")
	}
}

// TestSetupTestDB_WhenDockerAvailable verifies SetupTestDB returns a working *sql.DB.
// Skips when Docker is unavailable.
func TestSetupTestDB_WhenDockerAvailable(t *testing.T) {
	conn := SetupTestDB(t)
	if conn == nil {
		t.Fatal("SetupTestDB() should return non-nil connection")
	}
	if err := conn.Ping(); err != nil {
		t.Errorf("SetupTestDB() connection ping: %v", err)
	}
}

// TestSetupTestDatabase_WhenDockerAvailable verifies SetupTestDatabase returns *database.DB and GetConnection works.
// Skips when Docker is unavailable.
func TestSetupTestDatabase_WhenDockerAvailable(t *testing.T) {
	db := SetupTestDatabase(t)
	if db == nil {
		t.Fatal("SetupTestDatabase() should return non-nil DB")
	}
	conn := db.GetConnection()
	if conn == nil {
		t.Fatal("GetConnection() should return non-nil")
	}
	if err := conn.Ping(); err != nil {
		t.Errorf("SetupTestDatabase() connection ping: %v", err)
	}
}
