package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type colimaVMBackend struct{}

func (colimaVMBackend) Name() string {
	return "Colima"
}

func (colimaVMBackend) EnsureInstalled() error {
	return ensureColima()
}

func (colimaVMBackend) Status(cfg Config) (bool, error) {
	return commandSucceeded("colima", "status", "-p", cfg.VMProfile)
}

func (colimaVMBackend) CurrentMountLines(cfg Config) ([]string, error) {
	return currentColimaMountLines(cfg.VMProfile)
}

func (colimaVMBackend) Start(cfg Config, mounts []string) error {
	args := []string{
		"start", cfg.VMProfile,
		"--runtime", cfg.VMRuntime,
		"--vm-type", cfg.VMType,
		"--arch", cfg.VMArch,
		"--cpu", fmt.Sprintf("%d", cfg.VMCPU),
		"--memory", fmt.Sprintf("%d", cfg.VMMemory),
		"--disk", fmt.Sprintf("%d", cfg.VMDisk),
		"--mount-type", cfg.VMMountType,
	}
	if cfg.VMForwardSSHAgent {
		args = append(args, "--ssh-agent")
	}
	if cfg.VMNetworkAddress {
		args = append(args, "--network-address")
	}
	for _, mount := range mounts {
		args = append(args, "--mount", fmt.Sprintf("%s:w", mount))
	}
	fmt.Printf("Executing:\n  colima %s\n", shellQuoteArgs(args))
	return runCommand("colima", args...)
}

func (colimaVMBackend) Stop(cfg Config) error {
	return runCommand("colima", "stop", "-p", cfg.VMProfile)
}

func (colimaVMBackend) RunRemoteCommand(cfg Config, command string) error {
	return runCommand("colima", "ssh", "-p", cfg.VMProfile, "--", "/usr/bin/bash", "-lc", command)
}

func (colimaVMBackend) RunRemoteScript(cfg Config, script string, args []string) error {
	sshArgs := append([]string{
		"ssh", "-p", cfg.VMProfile, "--", "/usr/bin/bash", "-s", "--",
	}, args...)
	return runCommandInput(script, "colima", sshArgs...)
}

func currentColimaMountLines(profile string) ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	profileConfig := filepath.Join(home, ".colima", profile, "colima.yaml")
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

func desiredMountLines(harnessPath string, repoHosts []string) []string {
	mounts := []string{fmt.Sprintf("%s|rw", harnessPath)}
	for _, host := range repoHosts {
		mounts = append(mounts, fmt.Sprintf("%s|rw", host))
	}
	return normalizeMountLines(mounts)
}

func formatMountDiff(current []string, desired []string) []string {
	currentByLocation := groupMountsByLocation(current)
	desiredByLocation := groupMountsByLocation(desired)

	allLocations := make(map[string]struct{}, len(currentByLocation)+len(desiredByLocation))
	for location := range currentByLocation {
		allLocations[location] = struct{}{}
	}
	for location := range desiredByLocation {
		allLocations[location] = struct{}{}
	}

	locations := make([]string, 0, len(allLocations))
	for location := range allLocations {
		locations = append(locations, location)
	}
	sort.Strings(locations)

	var diff []string
	for _, location := range locations {
		currentMounts := currentByLocation[location]
		desiredMounts := desiredByLocation[location]

		currentSet := make(map[string]struct{}, len(currentMounts))
		desiredSet := make(map[string]struct{}, len(desiredMounts))
		for _, mount := range currentMounts {
			currentSet[mount] = struct{}{}
		}
		for _, mount := range desiredMounts {
			desiredSet[mount] = struct{}{}
		}

		for _, mount := range currentMounts {
			if _, ok := desiredSet[mount]; !ok {
				diff = append(diff, "- "+humanizeMountLine(mount))
			}
		}
		for _, mount := range desiredMounts {
			if _, ok := currentSet[mount]; !ok {
				diff = append(diff, "+ "+humanizeMountLine(mount))
			}
		}
	}

	return diff
}

func humanizeMountLine(mount string) string {
	parts := strings.SplitN(mount, "|", 2)
	if len(parts) != 2 {
		return mount
	}
	return fmt.Sprintf("%s (%s)", parts[0], parts[1])
}

func groupMountsByLocation(mounts []string) map[string][]string {
	grouped := make(map[string][]string)
	for _, mount := range mounts {
		location := mountLocation(mount)
		grouped[location] = append(grouped[location], mount)
	}
	for location := range grouped {
		sort.Strings(grouped[location])
	}
	return grouped
}

func mountLocation(mount string) string {
	return strings.SplitN(mount, "|", 2)[0]
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
