package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ensureCommand(name string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("%s is required but not installed", name)
	}
	return nil
}

func ensureColima() error {
	if err := ensureCommand("colima"); err != nil {
		return fmt.Errorf("colima is required for Harbour. Install it with: brew install colima: %w", err)
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
