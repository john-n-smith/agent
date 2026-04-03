package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var userConfigDir = os.UserConfigDir

type Config struct {
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
	return Config{
		ColimaProfile:         "harbour",
		ColimaRuntime:         "docker",
		ColimaVMType:          "vz",
		ColimaArch:            "aarch64",
		ColimaCPU:             4,
		ColimaMemory:          8,
		ColimaDisk:            100,
		ColimaMountType:       "virtiofs",
		ColimaForwardSSHAgent: true,
		ColimaNetworkAddress:  false,
		CodexVersion:          "latest",
		ClaudeCodeVersion:     "latest",
		HarnessPath:           "",
		WorkspaceRoot:         "",
		ActiveAgent:           "",
		DefaultCommand:        "agent",
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

	if cfg.ColimaCPU <= 0 {
		return fmt.Errorf("colima_cpu must be positive")
	}
	if cfg.ColimaMemory <= 0 {
		return fmt.Errorf("colima_memory must be positive")
	}
	if cfg.ColimaDisk <= 0 {
		return fmt.Errorf("colima_disk must be positive")
	}
	return nil
}
