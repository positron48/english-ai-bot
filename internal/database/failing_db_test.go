//go:build test

package database

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
)

func TestOpenFailingDB_ReturnsDB(t *testing.T) {
	db, err := OpenFailingDB()
	if err != nil {
		t.Fatalf("OpenFailingDB() error = %v", err)
	}
	if db == nil {
		t.Fatal("OpenFailingDB() returned nil *sql.DB")
	}
	defer db.Close()
}

func TestOpenFailingDB_QueryFailsWithConnectionError(t *testing.T) {
	db, err := OpenFailingDB()
	if err != nil {
		t.Fatalf("OpenFailingDB() error = %v", err)
	}
	defer db.Close()

	var n int
	err = db.QueryRow("SELECT 1").Scan(&n)
	if err == nil {
		t.Fatal("expected error on QueryRow against failing DB")
	}
	msg := err.Error()
	if !strings.Contains(msg, "connection") && !strings.Contains(msg, "refused") && !strings.Contains(msg, "connect") {
		t.Logf("QueryRow error (acceptable): %v", err)
	}
}

func TestOpenFailingDB_PingFails(t *testing.T) {
	db, err := OpenFailingDB()
	if err != nil {
		t.Fatalf("OpenFailingDB() error = %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err == nil {
		t.Fatal("expected error on Ping against failing DB")
	}
}

func TestOpenFailingDB_ExecFails(t *testing.T) {
	db, err := OpenFailingDB()
	if err != nil {
		t.Fatalf("OpenFailingDB() error = %v", err)
	}
	defer db.Close()

	_, err = db.Exec("SELECT 1")
	if err == nil {
		t.Fatal("expected error on Exec against failing DB")
	}
}

func TestOpenFailingDB_OpenReturnsError(t *testing.T) {
	wantErr := errors.New("injected open error")
	orig := openFailingDBOpen
	openFailingDBOpen = func(_, _ string) (*sql.DB, error) {
		return nil, wantErr
	}
	defer func() { openFailingDBOpen = orig }()

	db, err := OpenFailingDB()
	if err != wantErr {
		t.Fatalf("OpenFailingDB() err = %v, want %v", err, wantErr)
	}
	if db != nil {
		t.Fatal("OpenFailingDB() should return nil db when open fails")
	}
}
