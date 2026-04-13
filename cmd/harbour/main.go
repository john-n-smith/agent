package main

import (
	"fmt"
	"os"
)

var version = "dev"
var (
	runProvisionCommand = runProvision
	runShellCommand     = runShell
	runAgentCommand     = runAgent
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	command := ""
	if len(args) > 0 {
		command = args[0]
		args = args[1:]
	}

	if command == "" {
		exists, err := configExists()
		if err != nil {
			return err
		}
		if exists {
			cfg, err := loadConfig(false)
			if err == nil && canUseDefaultCommand(cfg) {
				command = cfg.DefaultCommand
			}
		}
	}

	switch command {
	case "", "help", "--help", "-h":
		if err := requireNoArgs(args); err != nil {
			return err
		}
		printUsage()
		return nil
	case "version", "--version", "-v":
		if err := requireNoArgs(args); err != nil {
			return err
		}
		fmt.Printf("harbour %s\n", version)
		return nil
	case "provision":
		if err := requireNoArgs(args); err != nil {
			return err
		}
		return runProvisionCommand()
	case "shell":
		if err := requireNoArgs(args); err != nil {
			return err
		}
		return runShellCommand()
	case "agent":
		if err := requireNoArgs(args); err != nil {
			return err
		}
		return runAgentCommand(false)
	case "yolo":
		if err := requireNoArgs(args); err != nil {
			return err
		}
		return runAgentCommand(true)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func printUsage() {
	fmt.Println("Usage: harbour [command]")
	fmt.Println()
	fmt.Println("Harbour provisions and runs an isolated Colima VM.")
	fmt.Println("Colima is required before running harbour provision.")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  provision   Provision the Harbour VM")
	fmt.Println("  shell       Open a shell in the Harbour VM")
	fmt.Println("  agent       Launch the provisioned agent")
	fmt.Println("  yolo        Launch the provisioned agent with relaxed permissions")
	fmt.Println("  help        Show this help")
	fmt.Println("  version     Show the Harbour version")
}

func requireNoArgs(args []string) error {
	if len(args) == 0 {
		return nil
	}
	return fmt.Errorf("unexpected arguments: %v", args)
}

func canUseDefaultCommand(cfg Config) bool {
	switch cfg.DefaultCommand {
	case "agent", "yolo":
		return cfg.WorkspacePath != "" && cfg.HarnessPath != "" && cfg.ActiveAgent != ""
	case "shell":
		return cfg.WorkspacePath != ""
	default:
		return false
	}
}
