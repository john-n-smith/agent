package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func colimaStatus(profile string) (bool, error) {
	return commandSucceeded("colima", "status", "-p", profile)
}

func currentMountLines(profile string) ([]string, error) {
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
	currentSet := make(map[string]struct{}, len(current))
	desiredSet := make(map[string]struct{}, len(desired))
	all := make(map[string]struct{}, len(current)+len(desired))

	for _, mount := range current {
		currentSet[mount] = struct{}{}
		all[mount] = struct{}{}
	}
	for _, mount := range desired {
		desiredSet[mount] = struct{}{}
		all[mount] = struct{}{}
	}

	keys := make([]string, 0, len(all))
	for mount := range all {
		keys = append(keys, mount)
	}
	sort.Strings(keys)

	var diff []string
	for _, mount := range keys {
		_, inCurrent := currentSet[mount]
		_, inDesired := desiredSet[mount]

		switch {
		case inCurrent && !inDesired:
			diff = append(diff, "- "+humanizeMountLine(mount))
		case !inCurrent && inDesired:
			diff = append(diff, "+ "+humanizeMountLine(mount))
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
