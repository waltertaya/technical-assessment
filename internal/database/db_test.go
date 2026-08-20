package database

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"gorm.io/gorm"
)

func TestConnectDBSuccess(t *testing.T) {
	originalOpen := openPostgres
	originalPrintln := logPrintln
	originalFatal := logFatal
	originalDB := DB

	t.Cleanup(func() {
		openPostgres = originalOpen
		logPrintln = originalPrintln
		logFatal = originalFatal
		DB = originalDB
	})

	t.Setenv("DATABASE_URL", "postgres://test-user:test-pass@localhost:5432/test_db")

	var gotDSN string
	var loggedMessage string
	expectedDB := &gorm.DB{}

	openPostgres = func(dsn string) (*gorm.DB, error) {
		gotDSN = dsn
		return expectedDB, nil
	}
	logPrintln = func(v ...any) {
		loggedMessage = fmt.Sprint(v...)
	}
	logFatal = func(v ...any) {
		t.Fatalf("did not expect fatal log, got: %v", v)
	}

	ConnectDB()

	if gotDSN != "postgres://test-user:test-pass@localhost:5432/test_db" {
		t.Fatalf("expected DATABASE_URL to be passed to openPostgres, got %q", gotDSN)
	}
	if DB != expectedDB {
		t.Fatalf("expected global DB to be assigned")
	}
	if loggedMessage != "Successfully connected to PostgreSQL database!" {
		t.Fatalf("unexpected success log: %q", loggedMessage)
	}
}

func TestConnectDBFailureLogsFatal(t *testing.T) {
	originalOpen := openPostgres
	originalPrintln := logPrintln
	originalFatal := logFatal
	originalDB := DB

	t.Cleanup(func() {
		openPostgres = originalOpen
		logPrintln = originalPrintln
		logFatal = originalFatal
		DB = originalDB
	})

	t.Setenv("DATABASE_URL", "postgres://test-user:test-pass@localhost:5432/test_db")

	expectedErr := errors.New("dial timeout")
	openPostgres = func(dsn string) (*gorm.DB, error) {
		return nil, expectedErr
	}
	logPrintln = func(v ...any) {
		t.Fatalf("did not expect success log, got: %v", v)
	}

	var fatalArgs []any
	logFatal = func(v ...any) {
		fatalArgs = v
		panic("fatal called")
	}

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("expected ConnectDB to call fatal logger")
		}
		if recovered != "fatal called" {
			t.Fatalf("unexpected panic value: %v", recovered)
		}
		if len(fatalArgs) != 1 {
			t.Fatalf("expected one fatal argument, got %d", len(fatalArgs))
		}

		err, ok := fatalArgs[0].(error)
		if !ok {
			t.Fatalf("expected fatal argument to be an error, got %T", fatalArgs[0])
		}
		msg := err.Error()
		if !strings.Contains(msg, "failed to open connection pool") {
			t.Fatalf("expected wrapped connection message, got %q", msg)
		}
		if !strings.Contains(msg, expectedErr.Error()) {
			t.Fatalf("expected original error in message, got %q", msg)
		}
	}()

	ConnectDB()
}
