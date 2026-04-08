package main

import "fmt"

type VMBackend interface {
	Name() string
	EnsureInstalled() error
	Status(cfg Config) (bool, error)
	CurrentMountLines(cfg Config) ([]string, error)
	Start(cfg Config, mounts []string) error
	Stop(cfg Config) error
	RunRemoteCommand(cfg Config, command string) error
	RunRemoteScript(cfg Config, script string, args []string) error
}

func resolveVMBackend(cfg Config) (VMBackend, error) {
	switch cfg.VMBackend {
	case "colima":
		return colimaVMBackend{}, nil
	default:
		return nil, fmt.Errorf("unsupported vm_backend=%s", cfg.VMBackend)
	}
}
