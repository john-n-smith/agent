package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/agent-harbour/harbour/cmd/harbour/vm"
)

import _ "embed"

//go:embed assets/provision_vm.sh
var provisionVMScript string

func runProvision() error {
	cfgPath, err := configPath()
	if err != nil {
		return err
	}
	fmt.Printf("Using config at %s\n\n", cfgPath)

	cfg, warning, err := loadConfigForProvision(cfgPath)
	if err != nil {
		return err
	}
	if warning != "" {
		fmt.Fprintf(os.Stderr, "Notice: invalid config: %s\n\n", warning)
	}
	vmBackend, err := vm.Resolve(cfg.vmConfig())
	if err != nil {
		return err
	}
	if err := vmBackend.EnsureInstalled(); err != nil {
		return err
	}

	workspacePromptDefault := cfg.WorkspacePath
	if workspacePromptDefault == "" {
		workspacePromptDefault = defaultWorkspacePromptPath()
	}
	reply, err := promptPathWithDefault("Workspace path: ", workspacePromptDefault)
	if err != nil {
		return err
	}
	if reply == "" {
		return fmt.Errorf("workspace_path is required")
	}
	cfg.WorkspacePath = reply
	cfg.WorkspacePath, err = canonicalPath(cfg.WorkspacePath)
	if err != nil {
		return err
	}
	if err := ensureDirectory(cfg.WorkspacePath, "workspace_path"); err != nil {
		return err
	}

	harnessPromptDefault := cfg.HarnessPath
	if harnessPromptDefault == "" {
		harnessPromptDefault = defaultHarnessPromptPath(reply)
	}
	reply, err = promptPathWithDefault(
		"Harness path: ",
		harnessPromptDefault,
	)
	if err != nil {
		return err
	}
	if reply == "" {
		return fmt.Errorf("harness_path is required")
	}
	cfg.HarnessPath = reply
	cfg.HarnessPath, err = canonicalPath(cfg.HarnessPath)
	if err != nil {
		return err
	}
	if err := ensureDirectory(cfg.HarnessPath, "harness_path"); err != nil {
		return err
	}
	if err := ensureSubdirectory(cfg.HarnessPath, cfg.WorkspacePath, "harness_path", "workspace_path"); err != nil {
		return err
	}
	fmt.Printf("harness_path=%s\n", cfg.HarnessPath)

	defaultAgent := "codex"
	if cfg.ActiveAgent != "" {
		defaultAgent = cfg.ActiveAgent
	}
	fmt.Printf("active_agent=%s\n", defaultAgent)

	defaultCommand := "agent"
	if cfg.DefaultCommand != "" {
		defaultCommand = cfg.DefaultCommand
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
			fmt.Printf("Resolving latest Codex release for Harbour profile %s\n", cfg.VMProfile)
		} else {
			fmt.Printf("Installing Codex %s in Harbour profile %s\n", requestedVersion, cfg.VMProfile)
		}
	case "claude":
		requestedVersion = cfg.ClaudeCodeVersion
		if requestedVersion == "latest" {
			fmt.Printf("Resolving latest Claude Code release for Harbour profile %s\n", cfg.VMProfile)
		} else {
			fmt.Printf("Installing Claude Code %s in Harbour profile %s\n", requestedVersion, cfg.VMProfile)
		}
	}

	mounts := []string{cfg.WorkspacePath}

	if err := os.Chdir(cfg.WorkspacePath); err != nil {
		return err
	}

	desiredMount := fmt.Sprintf("%s|rw", cfg.WorkspacePath)
	hasDesiredMount, err := vmBackend.HasExactMount(desiredMount)
	if err != nil {
		return err
	}

	running, err := vmBackend.Status()
	if err != nil {
		return err
	}
	if running {
		if !hasDesiredMount {
			fmt.Printf("Configured mounts differ from the running Harbour profile %s.\n", cfg.VMProfile)
			fmt.Printf("\nDesired mount:\n  %s\n", strings.Replace(desiredMount, "|", " (", 1)+")")
			ok, err := promptYesNo(fmt.Sprintf("\nRestart %s now to apply mount changes? [y/N] ", vmBackend.Name()))
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("aborted without restarting %s", vmBackend.Name())
			}
			if err := vmBackend.Stop(); err != nil {
				return err
			}
			if err := vmBackend.Start(mounts); err != nil {
				return err
			}
		} else {
			fmt.Printf("Harbour profile %s is already running.\n", cfg.VMProfile)
		}
	} else {
		if err := vmBackend.Start(mounts); err != nil {
			return err
		}
	}

	agentsPath := filepath.Join(cfg.HarnessPath, "AGENTS.md")
	if _, err := os.Stat(agentsPath); err != nil {
		return fmt.Errorf("%s is missing. Create it in harness_path before provisioning", agentsPath)
	}
	skillsPath := filepath.Join(cfg.HarnessPath, "skills")
	agentsData, err := os.ReadFile(agentsPath)
	if err != nil {
		return err
	}
	agentsB64 := base64.StdEncoding.EncodeToString(agentsData)

	hostUID := fmt.Sprintf("%d", os.Getuid())
	hostGID := fmt.Sprintf("%d", os.Getgid())
	scriptArgs := []string{
		selectedAgent,
		requestedVersion,
		agentsPath,
		skillsPath,
		agentsB64,
		hostUID,
		hostGID,
		cfg.WorkspacePath,
	}

	if err := vmBackend.RunRemoteScript(provisionVMScript, scriptArgs); err != nil {
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
		fmt.Printf("Provisioned Codex %s, linked %s/AGENTS.md, and synced custom skills.\n", requestedVersion, cfg.WorkspacePath)
	case "claude":
		fmt.Printf("Provisioned Claude Code %s, linked %s/CLAUDE.md, and synced custom skills.\n", requestedVersion, cfg.WorkspacePath)
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
	vmBackend, err := vm.Resolve(cfg.vmConfig())
	if err != nil {
		return err
	}
	if err := vmBackend.EnsureInstalled(); err != nil {
		return err
	}
	running, err := vmBackend.Status()
	if err != nil {
		return err
	}
	if !running {
		return fmt.Errorf("Harbour profile %s is not running. Start it with harbour provision", cfg.VMProfile)
	}
	fmt.Printf("Opening shell in Harbour profile %s\n", cfg.VMProfile)
	if err := os.Chdir(cfg.WorkspacePath); err != nil {
		return err
	}
	command := fmt.Sprintf("cd %q && exec /usr/bin/bash -l", cfg.WorkspacePath)
	return vmBackend.RunRemoteCommand(command)
}

func runAgent(yolo bool) error {
	cfg, configPath, err := requireProvisionedConfig(true)
	if err != nil {
		return err
	}
	vmBackend, err := vm.Resolve(cfg.vmConfig())
	if err != nil {
		return err
	}
	if err := vmBackend.EnsureInstalled(); err != nil {
		return err
	}
	running, err := vmBackend.Status()
	if err != nil {
		return err
	}
	if !running {
		return fmt.Errorf("Harbour profile %s is not running. Start it with harbour provision", cfg.VMProfile)
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

	fmt.Printf("Launching %s in Harbour profile %s\n", agentName, cfg.VMProfile)
	if err := os.Chdir(cfg.WorkspacePath); err != nil {
		return err
	}

	remoteScript := buildAgentRemoteScript(cfg, yolo, agentCommand, instructionFile)
	return vmBackend.RunRemoteCommand(remoteScript)
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
	if cfg.VMProfile == "" {
		return Config{}, "", fmt.Errorf("vm_profile is not set in %s. Run harbour provision", cfgPath)
	}
	if cfg.WorkspacePath == "" {
		return Config{}, "", fmt.Errorf("workspace_path is not set in %s. Run harbour provision", cfgPath)
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

workspace_path=%q
harbour_harness_dir=%q
yolo_mode=%q
agent_command=%q
instruction_file=%q
export PATH="${HOME}/.local/bin:${PATH}"

if ! command -v "${agent_command}" >/dev/null 2>&1; then
  echo "${agent_command} is not installed in the VM. Run harbour provision." >&2
  exit 127
fi

if [[ ! -d "${workspace_path}" ]]; then
  echo "${workspace_path} is not visible in the VM." >&2
  echo "Check the current mount layout, stop the Harbour VM, and run harbour provision again." >&2
  exit 127
fi

if [[ ! -f "${workspace_path}/${instruction_file}" ]]; then
  echo "${workspace_path}/${instruction_file} is missing in the VM." >&2
  echo "Run harbour provision." >&2
  exit 127
fi

if [[ ! -d "${harbour_harness_dir}" ]]; then
  echo "harness_path is not visible in the VM at ${harbour_harness_dir}." >&2
  echo "Keep the harness inside ${workspace_path}, stop the Harbour VM, and run harbour provision again." >&2
  exit 127
fi

cd "${workspace_path}"

case "${agent_command}" in
  codex)
    if [[ "${yolo_mode}" == "true" ]]; then
      args=(--dangerously-bypass-approvals-and-sandbox -C "${workspace_path}")
    else
      args=(--sandbox workspace-write -C "${workspace_path}")
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
`, cfg.WorkspacePath, cfg.HarnessPath, yoloMode, agentCommand, instructionFile)
}
