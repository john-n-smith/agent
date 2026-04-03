package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

import _ "embed"

//go:embed assets/provision_vm.sh
var provisionVMScript string

func runProvision() error {
	if err := ensureCommand("colima"); err != nil {
		return err
	}

	cfg, err := loadConfig(true)
	if err != nil {
		return err
	}
	cfgPath, err := configPath()
	if err != nil {
		return err
	}
	fmt.Printf("Using Harbour config %s\n", cfgPath)

	if cfg.HarnessPath == "" {
		reply, err := promptPath("Your Harbour harness path, e.g. ~/git/my-harbour-harness: ")
		if err != nil {
			return err
		}
		if reply == "" {
			return fmt.Errorf("harness_path is required")
		}
		cfg.HarnessPath = reply
	}
	cfg.HarnessPath, err = canonicalPath(cfg.HarnessPath)
	if err != nil {
		return err
	}
	if err := ensureDirectory(cfg.HarnessPath, "harness_path"); err != nil {
		return err
	}
	fmt.Printf("harness_path=%s\n", cfg.HarnessPath)

	reposFile := reposFilePath(cfg.HarnessPath)
	if _, err := os.Stat(reposFile); err != nil {
		return fmt.Errorf("%s is missing. Create it in harbour-harness", reposFile)
	}

	if cfg.WorkspaceRoot == "" {
		reply, err := promptPath("Workspace root, e.g. ~/git: ")
		if err != nil {
			return err
		}
		if reply == "" {
			return fmt.Errorf("workspace_root is required")
		}
		cfg.WorkspaceRoot = reply
	}
	cfg.WorkspaceRoot, err = canonicalPath(cfg.WorkspaceRoot)
	if err != nil {
		return err
	}
	if err := ensureDirectory(cfg.WorkspaceRoot, "workspace_root"); err != nil {
		return err
	}
	fmt.Printf("workspace_root=%s\n", cfg.WorkspaceRoot)

	defaultAgent := "codex"
	switch cfg.ActiveAgent {
	case "codex", "claude":
		defaultAgent = cfg.ActiveAgent
	case "":
	default:
		fmt.Fprintf(os.Stderr, "Ignoring invalid active_agent=%s in %s. Using %s.\n", cfg.ActiveAgent, cfgPath, defaultAgent)
	}
	fmt.Printf("active_agent=%s\n", defaultAgent)

	defaultCommand := "agent"
	switch cfg.DefaultCommand {
	case "agent", "shell", "yolo":
		defaultCommand = cfg.DefaultCommand
	case "":
	default:
		fmt.Fprintf(os.Stderr, "Ignoring invalid default_command=%s in %s. Using %s.\n", cfg.DefaultCommand, cfgPath, defaultCommand)
	}
	fmt.Printf("default_command=%s\n", defaultCommand)

	selectedAgent, err := promptChoice(
		fmt.Sprintf("Select the agent to provision [codex/claude] [%s]: ", defaultAgent),
		[]string{"codex", "claude"},
		defaultAgent,
	)
	if err != nil {
		return err
	}

	selectedDefaultCommand, err := promptChoice(
		fmt.Sprintf("Select the default harbour command [agent/yolo/shell] [%s]: ", defaultCommand),
		[]string{"agent", "yolo", "shell"},
		defaultCommand,
	)
	if err != nil {
		return err
	}

	requestedVersion := cfg.CodexVersion
	switch selectedAgent {
	case "codex":
		requestedVersion = cfg.CodexVersion
		if requestedVersion == "latest" {
			fmt.Printf("Resolving latest Codex release for Harbour profile %s\n", cfg.ColimaProfile)
		} else {
			fmt.Printf("Installing Codex %s in Harbour profile %s\n", requestedVersion, cfg.ColimaProfile)
		}
	case "claude":
		requestedVersion = cfg.ClaudeCodeVersion
		if requestedVersion == "latest" {
			fmt.Printf("Resolving latest Claude Code release for Harbour profile %s\n", cfg.ColimaProfile)
		} else {
			fmt.Printf("Installing Claude Code %s in Harbour profile %s\n", requestedVersion, cfg.ColimaProfile)
		}
	}

	repoHosts, err := existingRepoHosts(reposFile, cfg.WorkspaceRoot, true)
	if err != nil {
		return err
	}

	startArgs := []string{
		"start", cfg.ColimaProfile,
		"--runtime", cfg.ColimaRuntime,
		"--vm-type", cfg.ColimaVMType,
		"--arch", cfg.ColimaArch,
		"--cpu", fmt.Sprintf("%d", cfg.ColimaCPU),
		"--memory", fmt.Sprintf("%d", cfg.ColimaMemory),
		"--disk", fmt.Sprintf("%d", cfg.ColimaDisk),
		"--mount-type", cfg.ColimaMountType,
	}
	if cfg.ColimaForwardSSHAgent {
		startArgs = append(startArgs, "--ssh-agent")
	}
	if cfg.ColimaNetworkAddress {
		startArgs = append(startArgs, "--network-address")
	}
	startArgs = append(startArgs, "--mount", fmt.Sprintf("%s:w", cfg.HarnessPath))
	for _, host := range repoHosts {
		startArgs = append(startArgs, "--mount", fmt.Sprintf("%s:w", host))
	}

	desiredMounts := desiredMountLines(cfg.HarnessPath, repoHosts)
	currentMounts, err := currentMountLines(cfg.ColimaProfile)
	if err != nil {
		return err
	}

	running, err := colimaStatus(cfg.ColimaProfile)
	if err != nil {
		return err
	}
	if running {
		if !equalStringSlices(desiredMounts, currentMounts) {
			fmt.Printf("Configured mounts differ from the running Harbour profile %s.\n", cfg.ColimaProfile)
			fmt.Printf("\nMount diff:\n")
			for _, line := range formatMountDiff(currentMounts, desiredMounts) {
				fmt.Printf("  %s\n", line)
			}
			ok, err := promptYesNo("\nRestart Colima now to apply mount changes? [y/N] ")
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("aborted without restarting Colima")
			}
			if err := runCommand("colima", "stop", "-p", cfg.ColimaProfile); err != nil {
				return err
			}
			fmt.Printf("Executing:\n  colima %s\n", shellQuoteArgs(startArgs))
			if err := runCommand("colima", startArgs...); err != nil {
				return err
			}
		} else {
			fmt.Printf("Harbour profile %s is already running.\n", cfg.ColimaProfile)
		}
	} else {
		fmt.Printf("Executing:\n  colima %s\n", shellQuoteArgs(startArgs))
		if err := runCommand("colima", startArgs...); err != nil {
			return err
		}
	}

	agentsPath := filepath.Join(cfg.HarnessPath, "AGENTS.md")
	if _, err := os.Stat(agentsPath); err != nil {
		return fmt.Errorf("%s is missing. Create it in harbour-harness before provisioning", agentsPath)
	}
	skillsPath := filepath.Join(cfg.HarnessPath, "skills")
	agentsData, err := os.ReadFile(agentsPath)
	if err != nil {
		return err
	}
	agentsB64 := base64.StdEncoding.EncodeToString(agentsData)

	hostUID := fmt.Sprintf("%d", os.Getuid())
	hostGID := fmt.Sprintf("%d", os.Getgid())
	sshArgs := []string{
		"ssh", "-p", cfg.ColimaProfile, "--", "/usr/bin/bash", "-s", "--",
		selectedAgent,
		requestedVersion,
		agentsPath,
		skillsPath,
		agentsB64,
		hostUID,
		hostGID,
		cfg.WorkspaceRoot,
	}

	if err := os.Chdir(cfg.WorkspaceRoot); err != nil {
		return err
	}
	if err := runCommandInput(provisionVMScript, "colima", sshArgs...); err != nil {
		return err
	}

	cfg.ActiveAgent = selectedAgent
	cfg.DefaultCommand = selectedDefaultCommand
	if err := saveConfig(cfg); err != nil {
		return err
	}
	fmt.Printf("Saved Harbour config values to %s.\n", cfgPath)

	switch selectedAgent {
	case "codex":
		fmt.Printf("Provisioned Codex %s, linked %s/AGENTS.md, and synced custom skills.\n", requestedVersion, cfg.WorkspaceRoot)
	case "claude":
		fmt.Printf("Provisioned Claude Code %s, linked %s/CLAUDE.md, and synced custom skills.\n", requestedVersion, cfg.WorkspaceRoot)
	}
	fmt.Printf("Default command is harbour %s.\n", cfg.DefaultCommand)
	fmt.Println("Run harbour to use the default command, or run harbour agent, harbour yolo, or harbour shell explicitly.")
	return nil
}

func runShell() error {
	cfg, _, err := requireProvisionedConfig(false)
	if err != nil {
		return err
	}
	running, err := colimaStatus(cfg.ColimaProfile)
	if err != nil {
		return err
	}
	if !running {
		return fmt.Errorf("Harbour profile %s is not running. Start it with harbour provision", cfg.ColimaProfile)
	}
	fmt.Printf("Opening shell in Harbour profile %s\n", cfg.ColimaProfile)
	if err := os.Chdir(cfg.WorkspaceRoot); err != nil {
		return err
	}
	command := fmt.Sprintf("cd %q && exec /usr/bin/bash -l", cfg.WorkspaceRoot)
	return runCommand("colima", "ssh", "-p", cfg.ColimaProfile, "--", "/usr/bin/bash", "-lc", command)
}

func runAgent(yolo bool) error {
	cfg, configPath, err := requireProvisionedConfig(true)
	if err != nil {
		return err
	}
	running, err := colimaStatus(cfg.ColimaProfile)
	if err != nil {
		return err
	}
	if !running {
		return fmt.Errorf("Harbour profile %s is not running. Start it with harbour provision", cfg.ColimaProfile)
	}

	agentName := ""
	agentCommand := ""
	instructionFile := ""
	switch cfg.ActiveAgent {
	case "codex":
		agentName = "Codex"
		agentCommand = "codex"
		instructionFile = "AGENTS.md"
	case "claude":
		agentName = "Claude Code"
		agentCommand = "claude"
		instructionFile = "CLAUDE.md"
	default:
		return fmt.Errorf("unsupported active_agent=%s in %s. Run harbour provision and choose codex or claude", cfg.ActiveAgent, configPath)
	}

	fmt.Printf("Launching %s in Harbour profile %s\n", agentName, cfg.ColimaProfile)
	if err := os.Chdir(cfg.WorkspaceRoot); err != nil {
		return err
	}

	remoteScript := buildAgentRemoteScript(cfg, yolo, agentCommand, instructionFile)
	return runCommand("colima", "ssh", "-p", cfg.ColimaProfile, "--", "/usr/bin/bash", "-lc", remoteScript)
}

func requireProvisionedConfig(requireHarness bool) (Config, string, error) {
	cfg, err := loadConfig(false)
	if err != nil {
		return Config{}, "", err
	}
	cfgPath, err := configPath()
	if err != nil {
		return Config{}, "", err
	}
	if cfg.ColimaProfile == "" {
		return Config{}, "", fmt.Errorf("colima_profile is not set in %s. Run harbour provision", cfgPath)
	}
	if cfg.WorkspaceRoot == "" {
		return Config{}, "", fmt.Errorf("workspace_root is not set in %s. Run harbour provision", cfgPath)
	}
	if requireHarness && cfg.HarnessPath == "" {
		return Config{}, "", fmt.Errorf("harness_path is not set in %s. Run harbour provision", cfgPath)
	}
	return cfg, cfgPath, nil
}

func buildAgentRemoteScript(cfg Config, yolo bool, agentCommand string, instructionFile string) string {
	yoloMode := "false"
	if yolo {
		yoloMode = "true"
	}

	return fmt.Sprintf(`set -euo pipefail

workspace_root=%q
harbour_harness_dir=%q
yolo_mode=%q
agent_command=%q
instruction_file=%q
export PATH="${HOME}/.local/bin:${PATH}"

if ! command -v "${agent_command}" >/dev/null 2>&1; then
  echo "${agent_command} is not installed in the VM. Run harbour provision." >&2
  exit 127
fi

if [[ ! -d "${workspace_root}" ]]; then
  echo "${workspace_root} is not visible in the VM." >&2
  echo "Check the current mount layout, stop the Colima profile, and run harbour provision again." >&2
  exit 127
fi

if [[ ! -f "${workspace_root}/${instruction_file}" ]]; then
  echo "${workspace_root}/${instruction_file} is missing in the VM." >&2
  echo "Run harbour provision." >&2
  exit 127
fi

if [[ ! -d "${harbour_harness_dir}" ]]; then
  echo "harbour-harness is not mounted in the VM at ${harbour_harness_dir}." >&2
  echo "Check the sibling harbour-harness repo, stop the Colima profile, and run harbour provision again." >&2
  exit 127
fi

cd "${workspace_root}"

case "${agent_command}" in
  codex)
    if [[ "${yolo_mode}" == "true" ]]; then
      args=(--dangerously-bypass-approvals-and-sandbox -C "${workspace_root}")
    else
      args=(--sandbox workspace-write -C "${workspace_root}")
    fi
    ;;
  claude)
    if [[ "${yolo_mode}" == "true" ]]; then
      args=(--dangerously-skip-permissions)
    else
      args=(--permission-mode acceptEdits)
    fi
    ;;
esac

exec "${agent_command}" "${args[@]}"
`, cfg.WorkspaceRoot, cfg.HarnessPath, yoloMode, agentCommand, instructionFile)
}

func equalStringSlices(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func shellQuoteArgs(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		quoted = append(quoted, fmt.Sprintf("%q", arg))
	}
	return strings.Join(quoted, " ")
}
