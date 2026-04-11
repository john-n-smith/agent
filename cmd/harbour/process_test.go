package main

import (
	"strings"
	"testing"

	"github.com/agent-harbour/harbour/cmd/harbour/vm"
)

func TestColimaEnsureInstalledReturnsInstallGuidance(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	err := vm.Colima{}.EnsureInstalled()
	if err == nil {
		t.Fatal("EnsureInstalled() returned nil error")
	}
	if !strings.Contains(err.Error(), "brew install colima") {
		t.Fatalf("unexpected error: %v", err)
	}
}
