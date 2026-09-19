package main

import "testing"

func TestMigrationCommandRejectsInvalidInputBeforeConnecting(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "")
	for _, args := range [][]string{nil, {"reset"}, {"up", "extra"}, {"up"}} {
		if err := run(args); err == nil {
			t.Fatalf("expected error for %v", args)
		}
	}
}
