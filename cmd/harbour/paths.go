package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func expandHome(path string) (string, error) {
	if path == "" {
		return path, nil
	}
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			return home, nil
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

func canonicalPath(path string) (string, error) {
	expanded, err := expandHome(path)
	if err != nil {
		return "", err
	}
	if expanded == "" {
		return "", nil
	}

	info, err := os.Stat(expanded)
	if err == nil && info.IsDir() {
		resolved, err := filepath.EvalSymlinks(expanded)
		if err == nil {
			expanded = resolved
		}
		return filepath.Abs(expanded)
	}

	parent := filepath.Dir(expanded)
	parentAbs, err := filepath.Abs(parent)
	if err != nil {
		return "", err
	}
	resolvedParent, err := filepath.EvalSymlinks(parentAbs)
	if err == nil {
		parentAbs = resolvedParent
	}
	return filepath.Join(parentAbs, filepath.Base(expanded)), nil
}

func ensureDirectory(path string, label string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s does not exist: %s", label, path)
		}
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory: %s", label, path)
	}
	return nil
}

func ensureSubdirectory(path string, parent string, pathLabel string, parentLabel string) error {
	relativePath, err := filepath.Rel(parent, path)
	if err != nil {
		return err
	}
	if relativePath == "." {
		return fmt.Errorf("%s must be inside %s, not equal to it: %s", pathLabel, parentLabel, path)
	}
	if relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("%s must be inside %s: %s is outside %s", pathLabel, parentLabel, path, parent)
	}
	return nil
}
