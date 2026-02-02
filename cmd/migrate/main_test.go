package main

import (
	"strings"
	"testing"
)

// T083: Test that findMigrationsDir returns error when not found.
func TestFindMigrationsDir_NotFound_ReturnsError(t *testing.T) {
	// Change to a temp directory where migrations won't exist
	tmpDir := t.TempDir()
	originalWd := mustGetwd(t)
	mustChdir(t, tmpDir)
	defer mustChdir(t, originalWd)

	_, err := findMigrationsDir()
	if err == nil {
		t.Fatal("expected error when migrations directory not found, got nil")
	}

	// Error should indicate what's wrong
	errStr := err.Error()
	if !strings.Contains(errStr, "migration") {
		t.Errorf("error should mention migrations, got: %s", errStr)
	}
}

// TestFindMigrationsDir_Found verifies it finds the migrations directory.
func TestFindMigrationsDir_Found(t *testing.T) {
	// Run from repo root where migrations should exist
	// This test assumes it runs from the repo root
	dir, err := findMigrationsDir()
	if err != nil {
		t.Skipf("skipping: no migrations directory found (run from repo root): %v", err)
	}

	if dir == "" {
		t.Error("expected non-empty directory path")
	}

	if !strings.Contains(dir, "migrations") {
		t.Errorf("expected path to contain 'migrations', got: %s", dir)
	}
}

// Helper functions.
func mustGetwd(t *testing.T) string {
	t.Helper()
	wd, err := getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	return wd
}

func mustChdir(t *testing.T, dir string) {
	t.Helper()
	if err := chdir(dir); err != nil {
		t.Fatalf("failed to change directory to %s: %v", dir, err)
	}
}
