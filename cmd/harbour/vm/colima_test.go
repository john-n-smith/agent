package vm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestColimaHasExactMountMatchesSingleWritableMount(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	profileDir := filepath.Join(homeDir, ".colima", "harbour")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() returned error: %v", err)
	}

	config := "mounts:\n  - location: /workspace\n    writable: true\n"
	configPath := filepath.Join(profileDir, "colima.yaml")
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	ok, err := Colima{cfg: Config{Profile: "harbour"}}.HasExactMount("/workspace|rw")
	if err != nil {
		t.Fatalf("HasExactMount() returned error: %v", err)
	}
	if !ok {
		t.Fatal("HasExactMount() = false, want true")
	}
}

func TestColimaHasExactMountRejectsWrongModeOrExtraMounts(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	profileDir := filepath.Join(homeDir, ".colima", "harbour")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() returned error: %v", err)
	}

	config := "" +
		"mounts:\n" +
		"  - location: /workspace\n" +
		"    writable: false\n" +
		"  - location: /other\n" +
		"    writable: true\n"
	configPath := filepath.Join(profileDir, "colima.yaml")
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	ok, err := Colima{cfg: Config{Profile: "harbour"}}.HasExactMount("/workspace|rw")
	if err != nil {
		t.Fatalf("HasExactMount() returned error: %v", err)
	}
	if ok {
		t.Fatal("HasExactMount() = true, want false")
	}
}
