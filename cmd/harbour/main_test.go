package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
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
	cfg.WorkspaceRoot = "/tmp/workspace"
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
	cfg.WorkspaceRoot = "/tmp/workspace"
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

func TestLoadConfigMigratesLegacyColimaKeys(t *testing.T) {
	configDir := withTestConfigDir(t)
	path := filepath.Join(configDir, "harbour", "config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() returned error: %v", err)
	}

	legacyConfig := `{
  "colima_profile": "legacy-profile",
  "colima_runtime": "docker",
  "colima_vm_type": "qemu",
  "colima_arch": "x86_64",
  "colima_cpu": 2,
  "colima_memory": 4,
  "colima_disk": 20,
  "colima_mount_type": "sshfs",
  "colima_forward_ssh_agent": false,
  "colima_network_address": false,
  "codex_version": "latest",
  "claude_code_version": "latest",
  "harness_path": "/tmp/harness",
  "workspace_root": "/tmp/workspace",
  "active_agent": "codex",
  "default_command": "agent"
}
`
	if err := os.WriteFile(path, []byte(legacyConfig), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	previousInput := promptInput
	promptInput = bufio.NewReader(strings.NewReader("yes\n"))
	t.Cleanup(func() {
		promptInput = previousInput
	})

	got, err := loadConfig(false)
	if err != nil {
		t.Fatalf("loadConfig(false) returned error: %v", err)
	}

	want := defaultConfig()
	want.VMProfile = "legacy-profile"
	want.VMRuntime = "docker"
	want.VMType = "qemu"
	want.VMArch = "x86_64"
	want.VMCPU = 2
	want.VMMemory = 4
	want.VMDisk = 20
	want.VMMountType = "sshfs"
	want.VMForwardSSHAgent = false
	want.VMNetworkAddress = false
	want.HarnessPath = "/tmp/harness"
	want.WorkspaceRoot = "/tmp/workspace"
	want.ActiveAgent = "codex"
	want.DefaultCommand = "agent"

	if got != want {
		t.Fatalf("loadConfig(false) = %#v, want %#v", got, want)
	}

	rewritten, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() returned error: %v", err)
	}
	rewrittenText := string(rewritten)
	if !strings.Contains(rewrittenText, "\"vm_backend\": \"colima\"") {
		t.Fatalf("rewritten config did not contain vm_backend:\n%s", rewrittenText)
	}
	if strings.Contains(rewrittenText, "\"colima_profile\"") {
		t.Fatalf("rewritten config still contained legacy colima_* keys:\n%s", rewrittenText)
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
	if !strings.Contains(err.Error(), "active_agent must be codex, claude, or empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseRepoHosts(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	workspaceRoot := filepath.Join(t.TempDir(), "workspace")

	reposFile := writeReposFile(t, `
- host_path: /srv/absolute-repo
- host_path: relative/repo
- host_path: ~/home-repo # keep this comment ignored
`)

	hosts, err := parseRepoHosts(reposFile, workspaceRoot)
	if err != nil {
		t.Fatalf("parseRepoHosts() returned error: %v", err)
	}

	want := []string{
		"/srv/absolute-repo",
		filepath.Join(workspaceRoot, "relative/repo"),
		filepath.Join(homeDir, "home-repo"),
	}
	if !reflect.DeepEqual(hosts, want) {
		t.Fatalf("parseRepoHosts() = %#v, want %#v", hosts, want)
	}
}

func TestParseRepoHostsRequiresWorkspaceRootForRelativePaths(t *testing.T) {
	reposFile := writeReposFile(t, `
- host_path: relative/repo
`)

	_, err := parseRepoHosts(reposFile, "")
	if err == nil {
		t.Fatal("parseRepoHosts() returned nil error")
	}
	if !strings.Contains(err.Error(), "workspace_root is not set") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExistingRepoHostsSkipsMissingPathsWithWarning(t *testing.T) {
	workspaceRoot := t.TempDir()
	existingHost := filepath.Join(workspaceRoot, "existing")
	if err := os.MkdirAll(existingHost, 0o755); err != nil {
		t.Fatalf("MkdirAll() returned error: %v", err)
	}

	reposFile := writeReposFile(t, `
- host_path: existing
- host_path: missing
`)

	var hosts []string
	_, stderr := captureOutput(t, func() {
		var err error
		hosts, err = existingRepoHosts(reposFile, workspaceRoot, true)
		if err != nil {
			t.Fatalf("existingRepoHosts() returned error: %v", err)
		}
	})

	wantHosts := []string{existingHost}
	if !reflect.DeepEqual(hosts, wantHosts) {
		t.Fatalf("existingRepoHosts() = %#v, want %#v", hosts, wantHosts)
	}
	if !strings.Contains(stderr, "Warning: skipping missing repo mount") {
		t.Fatalf("stderr did not contain missing-path warning:\n%s", stderr)
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

func writeReposFile(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "repos.yaml")
	if err := os.WriteFile(path, []byte(strings.TrimSpace(contents)+"\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile() returned error: %v", err)
	}
	return path
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
