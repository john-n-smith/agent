SHELL := /bin/zsh
.DEFAULT_GOAL := help

PROJECT_ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))

.PHONY: help destroy provision shell agent yolo

help:
	@printf "Available targets:\n"
	@printf "  make destroy                    Stop the Colima VM profile used by this harness\n"
	@printf "  make provision                  Start the Colima VM, update Codex in the VM, link AGENTS.md at WORKSPACE_ROOT, and sync skills\n"
	@printf "  make shell                      Open an interactive shell in the Colima VM\n"
	@printf "  make agent                      Launch Codex inside the Colima VM\n"
	@printf "  make yolo                       Launch Codex with approvals and sandbox disabled\n"

destroy:
	@$(PROJECT_ROOT)/scripts/destroy

provision:
	@$(PROJECT_ROOT)/scripts/provision

shell:
	@$(PROJECT_ROOT)/scripts/shell

agent:
	@$(PROJECT_ROOT)/scripts/agent

yolo:
	@$(PROJECT_ROOT)/scripts/agent --yolo
