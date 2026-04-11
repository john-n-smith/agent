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
	WorkspaceRoot     string `json:"workspace_root"`
	ActiveAgent       string `json:"active_agent"`
	DefaultCommand    string `json:"default_command"`
}

type legacyConfig struct {
	ColimaProfile         string `json:"colima_profile"`
	ColimaRuntime         string `json:"colima_runtime"`
	ColimaVMType          string `json:"colima_vm_type"`
	ColimaArch            string `json:"colima_arch"`
	ColimaCPU             int    `json:"colima_cpu"`
	ColimaMemory          int    `json:"colima_memory"`
	ColimaDisk            int    `json:"colima_disk"`
	ColimaMountType       string `json:"colima_mount_type"`
	ColimaForwardSSHAgent bool   `json:"colima_forward_ssh_agent"`
	ColimaNetworkAddress  bool   `json:"colima_network_address"`
	CodexVersion          string `json:"codex_version"`
	ClaudeCodeVersion     string `json:"claude_code_version"`
	HarnessPath           string `json:"harness_path"`
	WorkspaceRoot         string `json:"workspace_root"`
	ActiveAgent           string `json:"active_agent"`
	DefaultCommand        string `json:"default_command"`
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
		WorkspaceRoot:     "",
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
	configErr := validateConfig(cfg)
	if configErr == nil {
		return cfg, nil
	}

	// If the file is not a valid vm_* config, try the legacy colima_* schema once and rewrite it.
	var legacy legacyConfig
	if err := json.Unmarshal(data, &legacy); err != nil {
		return Config{}, fmt.Errorf("invalid Harbour config %s: %w", path, err)
	}
	if !legacy.hasLegacyVMFields() {
		return Config{}, fmt.Errorf("invalid Harbour config %s: %w; no legacy colima_* fields found", path, configErr)
	}
	ok, err := promptYesNo("Migrate legacy Harbour config from colima_* keys to vm_* keys? [y/N] ")
	if err != nil {
		return Config{}, err
	}
	if !ok {
		return Config{}, fmt.Errorf("aborted without migrating legacy Harbour config")
	}
	migrated := legacy.toConfig()
	if err := validateConfig(migrated); err != nil {
		return Config{}, fmt.Errorf("invalid Harbour config %s: %w", path, err)
	}
	if err := saveConfig(migrated); err != nil {
		return Config{}, err
	}
	return migrated, nil
}

func (legacy legacyConfig) toConfig() Config {
	cfg := defaultConfig()
	cfg.VMProfile = legacy.ColimaProfile
	cfg.VMRuntime = legacy.ColimaRuntime
	cfg.VMType = legacy.ColimaVMType
	cfg.VMArch = legacy.ColimaArch
	cfg.VMCPU = legacy.ColimaCPU
	cfg.VMMemory = legacy.ColimaMemory
	cfg.VMDisk = legacy.ColimaDisk
	cfg.VMMountType = legacy.ColimaMountType
	cfg.VMForwardSSHAgent = legacy.ColimaForwardSSHAgent
	cfg.VMNetworkAddress = legacy.ColimaNetworkAddress
	cfg.CodexVersion = legacy.CodexVersion
	cfg.ClaudeCodeVersion = legacy.ClaudeCodeVersion
	cfg.HarnessPath = legacy.HarnessPath
	cfg.WorkspaceRoot = legacy.WorkspaceRoot
	cfg.ActiveAgent = legacy.ActiveAgent
	cfg.DefaultCommand = legacy.DefaultCommand
	return cfg
}

func (legacy legacyConfig) hasLegacyVMFields() bool {
	return legacy.ColimaProfile != "" ||
		legacy.ColimaRuntime != "" ||
		legacy.ColimaVMType != "" ||
		legacy.ColimaArch != "" ||
		legacy.ColimaCPU > 0 ||
		legacy.ColimaMemory > 0 ||
		legacy.ColimaDisk > 0 ||
		legacy.ColimaMountType != ""
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
		return fmt.Errorf("vm_backend must be colima")
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
