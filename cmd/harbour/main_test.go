package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWithoutConfigShowsHelp(t *testing.T) {
	withTestConfigDir(t)

	stdout, stderr := captureOutput(t, func() {
		if err := run(nil); err != nil {
			t.Fatalf("run(nil) returned error: %v", err)
		}
	})

	if !strings.Contains(stdout, "Usage: harbour [command]") {
		t.Fatalf("stdout did not contain help output:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Colima is required before running harbour provision.") {
		t.Fatalf("stdout did not contain Colima prerequisite guidance:\n%s", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr was not empty:\n%s", stderr)
	}
}

func TestRunUsesConfiguredDefaultCommand(t *testing.T) {
	withTestConfigDir(t)

	cfg := defaultConfig()
	cfg.DefaultCommand = "agent"
	cfg.ActiveAgent = "codex"
	cfg.HarnessPath = "/tmp/harness"
	cfg.WorkspacePath = "/tmp/workspace"
	if err := saveConfig(cfg); err != nil {
		t.Fatalf("saveConfig() returned error: %v", err)
	}

	called := false
	restore := stubRunAgent(func(yolo bool) error {
		called = true
		if yolo {
			t.Fatalf("runAgent called in yolo mode")
		}
		return nil
	})
	defer restore()

	stdout, stderr := captureOutput(t, func() {
		if err := run(nil); err != nil {
			t.Fatalf("run(nil) returned error: %v", err)
		}
	})

	if !called {
		t.Fatal("runAgent was not called")
	}
	if stdout != "" {
		t.Fatalf("stdout was not empty:\n%s", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr was not empty:\n%s", stderr)
	}
}

func TestRunFallsBackToHelpWhenConfigIsInvalid(t *testing.T) {
	configDir := withTestConfigDir(t)
	configPath := filepath.Join(configDir, "harbour", "config.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() returned error: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("{\"vm_backend\":\"\"}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	restore := stubRunAgent(func(yolo bool) error {
		t.Fatal("runAgent should not be called for invalid config")
		return nil
	})
	defer restore()

	stdout, stderr := captureOutput(t, func() {
		if err := run(nil); err != nil {
			t.Fatalf("run(nil) returned error: %v", err)
		}
	})

	if !strings.Contains(stdout, "Usage: harbour [command]") {
		t.Fatalf("stdout did not contain help output:\n%s", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr was not empty:\n%s", stderr)
	}
}

func TestRunFallsBackToHelpWhenConfigIsIncompleteForDefaultCommand(t *testing.T) {
	withTestConfigDir(t)

	cfg := defaultConfig()
	cfg.DefaultCommand = "agent"
	if err := saveConfig(cfg); err != nil {
		t.Fatalf("saveConfig() returned error: %v", err)
	}

	restore := stubRunAgent(func(yolo bool) error {
		t.Fatal("runAgent should not be called for incomplete config")
		return nil
	})
	defer restore()

	stdout, stderr := captureOutput(t, func() {
		if err := run(nil); err != nil {
			t.Fatalf("run(nil) returned error: %v", err)
		}
	})

	if !strings.Contains(stdout, "Usage: harbour [command]") {
		t.Fatalf("stdout did not contain help output:\n%s", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr was not empty:\n%s", stderr)
	}
}

func TestRunProvisionRecoversFromInvalidConfig(t *testing.T) {
	configDir := withTestConfigDir(t)
	configPath := filepath.Join(configDir, "harbour", "config.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll() returned error: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("{\"vm_backend\":\"\"}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	called := false
	previous := runProvisionCommand
	runProvisionCommand = func() error {
		called = true
		return nil
	}
	t.Cleanup(func() {
		runProvisionCommand = previous
	})

	stdout, stderr := captureOutput(t, func() {
		if err := run([]string{"provision"}); err != nil {
			t.Fatalf("run(provision) returned error: %v", err)
		}
	})

	if !called {
		t.Fatal("runProvision was not called")
	}
	if stdout != "" {
		t.Fatalf("stdout was not empty:\n%s", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr was not empty:\n%s", stderr)
	}
}

func TestExplicitHelpBypassesConfiguredDefaultCommand(t *testing.T) {
	withTestConfigDir(t)

	cfg := defaultConfig()
	cfg.DefaultCommand = "agent"
	if err := saveConfig(cfg); err != nil {
		t.Fatalf("saveConfig() returned error: %v", err)
	}

	restore := stubRunAgent(func(yolo bool) error {
		t.Fatal("runAgent should not be called for explicit help")
		return nil
	})
	defer restore()

	stdout, stderr := captureOutput(t, func() {
		if err := run([]string{"help"}); err != nil {
			t.Fatalf("run(help) returned error: %v", err)
		}
	})

	if !strings.Contains(stdout, "Commands:") {
		t.Fatalf("stdout did not contain help output:\n%s", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr was not empty:\n%s", stderr)
	}
}

func TestExplicitVersionBypassesConfiguredDefaultCommand(t *testing.T) {
	withTestConfigDir(t)

	cfg := defaultConfig()
	cfg.DefaultCommand = "agent"
	if err := saveConfig(cfg); err != nil {
		t.Fatalf("saveConfig() returned error: %v", err)
	}

	restore := stubRunAgent(func(yolo bool) error {
		t.Fatal("runAgent should not be called for explicit version")
		return nil
	})
	defer restore()

	previousVersion := version
	version = "test-version"
	t.Cleanup(func() {
		version = previousVersion
	})

	stdout, stderr := captureOutput(t, func() {
		if err := run([]string{"version"}); err != nil {
			t.Fatalf("run(version) returned error: %v", err)
		}
	})

	if stdout != "harbour test-version\n" {
		t.Fatalf("stdout = %q, want %q", stdout, "harbour test-version\n")
	}
	if stderr != "" {
		t.Fatalf("stderr was not empty:\n%s", stderr)
	}
}

func TestLoadConfigCreatesDefaultConfig(t *testing.T) {
	configDir := withTestConfigDir(t)

	cfg, err := loadConfig(true)
	if err != nil {
		t.Fatalf("loadConfig(true) returned error: %v", err)
	}

	if cfg != defaultConfig() {
		t.Fatalf("loadConfig(true) = %#v, want %#v", cfg, defaultConfig())
	}

	path := filepath.Join(configDir, "harbour", "config.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("config file was not created: %v", err)
	}
}

func TestApplyPlatformDefaults(t *testing.T) {
	tests := []struct {
		name      string
		goos      string
		goarch    string
		wantVM    string
		wantArch  string
		wantMount string
	}{
		{
			name:      "Intel macOS",
			goos:      "darwin",
			goarch:    "amd64",
			wantVM:    "qemu",
			wantArch:  "x86_64",
			wantMount: "sshfs",
		},
		{
			name:      "Apple Silicon macOS",
			goos:      "darwin",
			goarch:    "arm64",
			wantVM:    "vz",
			wantArch:  "aarch64",
			wantMount: "virtiofs",
		},
		{
			name:      "Linux amd64",
			goos:      "linux",
			goarch:    "amd64",
			wantVM:    "vz",
			wantArch:  "aarch64",
			wantMount: "virtiofs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultConfig()
			cfg.VMType = "vz"
			cfg.VMArch = "aarch64"
			cfg.VMMountType = "virtiofs"

			applyPlatformDefaults(&cfg, tt.goos, tt.goarch)

			if cfg.VMType != tt.wantVM {
				t.Fatalf("applyPlatformDefaults().VMType = %q, want %q", cfg.VMType, tt.wantVM)
			}
			if cfg.VMArch != tt.wantArch {
				t.Fatalf("applyPlatformDefaults().VMArch = %q, want %q", cfg.VMArch, tt.wantArch)
			}
			if cfg.VMMountType != tt.wantMount {
				t.Fatalf("applyPlatformDefaults().VMMountType = %q, want %q", cfg.VMMountType, tt.wantMount)
			}
		})
	}
}

func TestSaveConfigRoundTrip(t *testing.T) {
	withTestConfigDir(t)

	cfg := defaultConfig()
	cfg.HarnessPath = "/tmp/harness"
	cfg.WorkspacePath = "/tmp/workspace"
	cfg.ActiveAgent = "claude"
	cfg.DefaultCommand = "shell"
	cfg.VMCPU = 8

	if err := saveConfig(cfg); err != nil {
		t.Fatalf("saveConfig() returned error: %v", err)
	}

	loaded, err := loadConfig(false)
	if err != nil {
		t.Fatalf("loadConfig(false) returned error: %v", err)
	}

	if loaded != cfg {
		t.Fatalf("loadConfig(false) = %#v, want %#v", loaded, cfg)
	}
}

func TestSaveConfigRejectsInvalidValues(t *testing.T) {
	withTestConfigDir(t)

	cfg := defaultConfig()
	cfg.ActiveAgent = "invalid"

	err := saveConfig(cfg)
	if err == nil {
		t.Fatal("saveConfig() returned nil error for invalid config")
	}
	if !strings.Contains(err.Error(), "`active_agent` must be codex, claude, or empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSaveConfigRejectsEmptyVMProfile(t *testing.T) {
	withTestConfigDir(t)

	cfg := defaultConfig()
	cfg.VMProfile = ""

	err := saveConfig(cfg)
	if err == nil {
		t.Fatal("saveConfig() returned nil error for invalid config")
	}
	if !strings.Contains(err.Error(), "`vm_profile` must not be empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnsureSubdirectoryAcceptsChildDirectory(t *testing.T) {
	workspacePath := t.TempDir()
	harnessPath := filepath.Join(workspacePath, ".harbour-harness")

	if err := os.MkdirAll(harnessPath, 0o755); err != nil {
		t.Fatalf("MkdirAll() returned error: %v", err)
	}

	if err := ensureSubdirectory(harnessPath, workspacePath, "harness_path", "workspace_path"); err != nil {
		t.Fatalf("ensureSubdirectory() returned error: %v", err)
	}
}

func TestEnsureSubdirectoryRejectsWorkspacePath(t *testing.T) {
	workspacePath := t.TempDir()

	err := ensureSubdirectory(workspacePath, workspacePath, "harness_path", "workspace_path")
	if err == nil {
		t.Fatal("ensureSubdirectory() returned nil error")
	}
	if !strings.Contains(err.Error(), "must be inside workspace_path, not equal to it") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnsureSubdirectoryRejectsPathOutsideWorkspacePath(t *testing.T) {
	workspacePath := t.TempDir()
	harnessPath := t.TempDir()

	err := ensureSubdirectory(harnessPath, workspacePath, "harness_path", "workspace_path")
	if err == nil {
		t.Fatal("ensureSubdirectory() returned nil error")
	}
	if !strings.Contains(err.Error(), "must be inside workspace_path") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func withTestConfigDir(t *testing.T) string {
	t.Helper()

	configDir := t.TempDir()
	previous := userConfigDir
	userConfigDir = func() (string, error) {
		return configDir, nil
	}
	t.Cleanup(func() {
		userConfigDir = previous
	})
	return configDir
}

func stubRunAgent(fn func(bool) error) func() {
	previous := runAgentCommand
	runAgentCommand = fn
	return func() {
		runAgentCommand = previous
	}
}

func captureOutput(t *testing.T, fn func()) (string, string) {
	t.Helper()

	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() returned error: %v", err)
	}
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() returned error: %v", err)
	}

	originalStdout := os.Stdout
	originalStderr := os.Stderr
	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	stdoutCh := make(chan string, 1)
	stderrCh := make(chan string, 1)
	go readPipeOutput(stdoutReader, stdoutCh)
	go readPipeOutput(stderrReader, stderrCh)

	fn()

	if err := stdoutWriter.Close(); err != nil {
		t.Fatalf("stdoutWriter.Close() returned error: %v", err)
	}
	if err := stderrWriter.Close(); err != nil {
		t.Fatalf("stderrWriter.Close() returned error: %v", err)
	}
	os.Stdout = originalStdout
	os.Stderr = originalStderr

	stdout := <-stdoutCh
	stderr := <-stderrCh

	if err := stdoutReader.Close(); err != nil {
		t.Fatalf("stdoutReader.Close() returned error: %v", err)
	}
	if err := stderrReader.Close(); err != nil {
		t.Fatalf("stderrReader.Close() returned error: %v", err)
	}

	return stdout, stderr
}

func readPipeOutput(reader *os.File, outputCh chan<- string) {
	var buffer bytes.Buffer
	_, _ = io.Copy(&buffer, reader)
	outputCh <- buffer.String()
}

func TestMain(m *testing.M) {
	status := m.Run()
	runProvisionCommand = runProvision
	runShellCommand = runShell
	runAgentCommand = runAgent
	userConfigDir = os.UserConfigDir
	os.Exit(status)
}
