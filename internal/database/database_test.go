package database

import (
	"context"
	"strings"
	"testing"
)

func TestConnectionErrorsDoNotExposeDSN(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	const dsn = "postgres://user:canary-secret@127.0.0.1:1/db?connect_timeout=1"
	if _, err := Open(ctx, dsn); err == nil || strings.Contains(err.Error(), "canary-secret") {
		t.Fatalf("unsafe connection error: %v", err)
	}
	if _, err := NewMigrator(ctx, dsn); err == nil || strings.Contains(err.Error(), "canary-secret") {
		t.Fatalf("unsafe migration error: %v", err)
	}
	if _, err := NewMigrator(ctx, ""); err == nil {
		t.Fatal("missing DSN accepted")
	}
}
