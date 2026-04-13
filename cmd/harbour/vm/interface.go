package vm

import "fmt"

type Config struct {
	Backend         string
	Profile         string
	Runtime         string
	Type            string
	Arch            string
	CPU             int
	Memory          int
	Disk            int
	MountType       string
	ForwardSSHAgent bool
	NetworkAddress  bool
}

type Backend interface {
	Name() string
	EnsureInstalled() error
	Status() (bool, error)
	HasExactMount(mount string) (bool, error)
	Start(mounts []string) error
	Stop() error
	RunRemoteCommand(command string) error
	RunRemoteScript(script string, args []string) error
}

func Resolve(cfg Config) (Backend, error) {
	switch cfg.Backend {
	case "colima":
		return Colima{cfg: cfg}, nil
	default:
		return nil, fmt.Errorf("Unsupported `vm_backend`=%q (supported: colima)", cfg.Backend)
	}
}
