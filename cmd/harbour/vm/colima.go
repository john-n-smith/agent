package vm

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type ColimaBackend struct {
	cfg Config
}

var _ Backend = ColimaBackend{}

func (ColimaBackend) Name() string {
	return "Colima"
}

func (ColimaBackend) EnsureInstalled() error {
	if err := ensureCommand("colima"); err != nil {
		return fmt.Errorf("colima is required for Harbour. Install it with: brew install colima: %w", err)
	}
	return nil
}

func (b ColimaBackend) Status() (bool, error) {
	return commandSucceeded("colima", "status", "-p", b.cfg.Profile)
}

func (b ColimaBackend) CurrentMountLines() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	profileConfig := filepath.Join(home, ".colima", b.cfg.Profile, "colima.yaml")
	file, err := os.Open(profileConfig)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var mounts []string
	scanner := bufio.NewScanner(file)
	inMounts := false
	location := ""
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "mounts:") {
			inMounts = true
			continue
		}
		if inMounts && trimmed != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			inMounts = false
		}
		if !inMounts {
			continue
		}
		if strings.HasPrefix(trimmed, "- location:") {
			location = strings.TrimSpace(strings.TrimPrefix(trimmed, "- location:"))
			continue
		}
		if strings.HasPrefix(trimmed, "writable:") && location != "" {
			mode := "ro"
			if strings.TrimSpace(strings.TrimPrefix(trimmed, "writable:")) == "true" {
				mode = "rw"
			}
			mounts = append(mounts, fmt.Sprintf("%s|%s", location, mode))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return normalizeMountLines(mounts), nil
}

func (b ColimaBackend) Start(mounts []string) error {
	args := []string{
		"start", b.cfg.Profile,
		"--runtime", b.cfg.Runtime,
		"--vm-type", b.cfg.Type,
		"--arch", b.cfg.Arch,
		"--cpu", fmt.Sprintf("%d", b.cfg.CPU),
		"--memory", fmt.Sprintf("%d", b.cfg.Memory),
		"--disk", fmt.Sprintf("%d", b.cfg.Disk),
		"--mount-type", b.cfg.MountType,
	}
	if b.cfg.ForwardSSHAgent {
		args = append(args, "--ssh-agent")
	}
	if b.cfg.NetworkAddress {
		args = append(args, "--network-address")
	}
	for _, mount := range mounts {
		args = append(args, "--mount", fmt.Sprintf("%s:w", mount))
	}
	fmt.Printf("Executing:\n  colima %s\n", shellQuoteArgs(args))
	return runCommand("colima", args...)
}

func (b ColimaBackend) Stop() error {
	return runCommand("colima", "stop", "-p", b.cfg.Profile)
}

func (b ColimaBackend) RunRemoteCommand(command string) error {
	return runCommand("colima", "ssh", "-p", b.cfg.Profile, "--", "/usr/bin/bash", "-lc", command)
}

func (b ColimaBackend) RunRemoteScript(script string, args []string) error {
	sshArgs := append([]string{
		"ssh", "-p", b.cfg.Profile, "--", "/usr/bin/bash", "-s", "--",
	}, args...)
	return runCommandInput(script, "colima", sshArgs...)
}

func ensureCommand(name string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("%s is required but not installed", name)
	}
	return nil
}

func runCommand(name string, args ...string) error {
	if err := ensureCommand(name); err != nil {
		return err
	}
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runCommandInput(input string, name string, args ...string) error {
	if err := ensureCommand(name); err != nil {
		return err
	}
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(input)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func commandSucceeded(name string, args ...string) (bool, error) {
	if err := ensureCommand(name); err != nil {
		return false, err
	}
	cmd := exec.Command(name, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	if _, ok := err.(*exec.ExitError); ok {
		return false, nil
	}
	return false, err
}

func normalizeMountLines(mounts []string) []string {
	if len(mounts) == 0 {
		return nil
	}

	sorted := append([]string(nil), mounts...)
	sort.Strings(sorted)

	normalized := sorted[:0]
	for _, mount := range sorted {
		if len(normalized) > 0 && normalized[len(normalized)-1] == mount {
			continue
		}
		normalized = append(normalized, mount)
	}
	return normalized
}

func shellQuoteArgs(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		quoted = append(quoted, fmt.Sprintf("%q", arg))
	}
	return strings.Join(quoted, " ")
}
