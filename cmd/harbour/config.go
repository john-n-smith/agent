package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/agent-harbour/harbour/cmd/harbour/vm"
)

var userConfigDir = os.UserConfigDir

type Config struct {
	VMBackend         string `json:"vm_backend"`
	VMProfile         string `json:"vm_profile"`
	VMRuntime         string `json:"vm_runtime"`
	VMType            string `json:"vm_type"`
	VMArch            string `json:"vm_arch"`
	VMCPU             int    `json:"vm_cpu"`
	VMMemory          int    `json:"vm_memory"`
	VMDisk            int    `json:"vm_disk"`
	VMMountType       string `json:"vm_mount_type"`
	VMForwardSSHAgent bool   `json:"vm_forward_ssh_agent"`
	VMNetworkAddress  bool   `json:"vm_network_address"`
	CodexVersion      string `json:"codex_version"`
	ClaudeCodeVersion string `json:"claude_code_version"`
	HarnessPath       string `json:"harness_path"`
	WorkspacePath     string `json:"workspace_path"`
	ActiveAgent       string `json:"active_agent"`
	DefaultCommand    string `json:"default_command"`
}

func defaultConfig() Config {
	cfg := Config{
		VMBackend:         "colima",
		VMProfile:         "harbour",
		VMRuntime:         "docker",
		VMType:            "vz",
		VMArch:            "aarch64",
		VMCPU:             4,
		VMMemory:          8,
		VMDisk:            100,
		VMMountType:       "virtiofs",
		VMForwardSSHAgent: true,
		VMNetworkAddress:  false,
		CodexVersion:      "latest",
		ClaudeCodeVersion: "latest",
		HarnessPath:       "",
		WorkspacePath:     "",
		ActiveAgent:       "",
		DefaultCommand:    "agent",
	}

	applyPlatformDefaults(&cfg, runtime.GOOS, runtime.GOARCH)
	return cfg
}

func applyPlatformDefaults(cfg *Config, goos string, goarch string) {
	if goos == "darwin" && goarch == "amd64" {
		cfg.VMType = "qemu"
		cfg.VMArch = "x86_64"
		cfg.VMMountType = "sshfs"
	}
}

func configPath() (string, error) {
	configDir, err := userConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "harbour", "config.json"), nil
}

func configExists() (bool, error) {
	path, err := configPath()
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func loadConfig(create bool) (Config, error) {
	path, err := configPath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := defaultConfig()
			if create {
				if err := saveConfig(cfg); err != nil {
					return Config{}, err
				}
			}
			return cfg, nil
		}
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("invalid Harbour config %s: %w", path, err)
	}
	if err := validateConfig(cfg); err != nil {
		return Config{}, fmt.Errorf("invalid Harbour config %s: %w", path, err)
	}
	return cfg, nil
}

func saveConfig(cfg Config) error {
	if err := validateConfig(cfg); err != nil {
		return err
	}

	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(filepath.Dir(path), "config-*.json")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func validateConfig(cfg Config) error {
	switch cfg.VMBackend {
	case "colima":
	default:
		return fmt.Errorf("unsupported vm_backend=%q (supported: colima)", cfg.VMBackend)
	}
	if cfg.VMProfile == "" {
		return fmt.Errorf("vm_profile must not be empty")
	}
	if cfg.VMRuntime == "" {
		return fmt.Errorf("vm_runtime must not be empty")
	}
	if cfg.VMType == "" {
		return fmt.Errorf("vm_type must not be empty")
	}
	if cfg.VMArch == "" {
		return fmt.Errorf("vm_arch must not be empty")
	}
	if cfg.VMMountType == "" {
		return fmt.Errorf("vm_mount_type must not be empty")
	}

	switch cfg.ActiveAgent {
	case "", "codex", "claude":
	default:
		return fmt.Errorf("active_agent must be codex, claude, or empty")
	}

	switch cfg.DefaultCommand {
	case "", "agent", "shell", "yolo":
	default:
		return fmt.Errorf("default_command must be agent, shell, yolo, or empty")
	}

	if cfg.VMCPU <= 0 {
		return fmt.Errorf("vm_cpu must be positive")
	}
	if cfg.VMMemory <= 0 {
		return fmt.Errorf("vm_memory must be positive")
	}
	if cfg.VMDisk <= 0 {
		return fmt.Errorf("vm_disk must be positive")
	}
	return nil
}

func (cfg Config) vmConfig() vm.Config {
	return vm.Config{
		Backend:         cfg.VMBackend,
		Profile:         cfg.VMProfile,
		Runtime:         cfg.VMRuntime,
		Type:            cfg.VMType,
		Arch:            cfg.VMArch,
		CPU:             cfg.VMCPU,
		Memory:          cfg.VMMemory,
		Disk:            cfg.VMDisk,
		MountType:       cfg.VMMountType,
		ForwardSSHAgent: cfg.VMForwardSSHAgent,
		NetworkAddress:  cfg.VMNetworkAddress,
	}
}
