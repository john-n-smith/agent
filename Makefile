SHELL := /bin/zsh
.DEFAULT_GOAL := help

PROJECT_ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))

.PHONY: help provision shell agent yolo

help:
	@printf "Available targets:\n"
	@printf "  make provision                  Interactively provision the environment\n"
	@printf "  make shell                      Open a shell in the Colima VM\n"
	@printf "  make agent                      Launch the agent inside the Colima VM\n"
	@printf "  make yolo                       Launch the agent with relaxed permissions\n"

provision:
	@$(PROJECT_ROOT)/scripts/provision

shell:
	@$(PROJECT_ROOT)/scripts/shell

agent:
	@$(PROJECT_ROOT)/scripts/agent

yolo:
	@$(PROJECT_ROOT)/scripts/agent --yolo
