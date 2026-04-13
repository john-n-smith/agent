package vm

import (
	"strings"
	"testing"
)

func TestResolveReturnsColimaBackend(t *testing.T) {
	cfg := Config{
		Backend: "colima",
		Profile: "harbour",
	}

	backend, err := Resolve(cfg)
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	colima, ok := backend.(Colima)
	if !ok {
		t.Fatalf("Resolve() returned %T, want vm.Colima", backend)
	}
	if colima.cfg != cfg {
		t.Fatalf("resolved backend cfg = %#v, want %#v", colima.cfg, cfg)
	}
}

func TestResolveRejectsUnsupportedBackend(t *testing.T) {
	_, err := Resolve(Config{Backend: "lima"})
	if err == nil {
		t.Fatal("Resolve() returned nil error")
	}
	if !strings.Contains(err.Error(), "Unsupported `vm_backend`=\"lima\" (supported: colima)") {
		t.Fatalf("unexpected error: %v", err)
	}
}
