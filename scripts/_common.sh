#!/bin/zsh

set -euo pipefail

SCRIPT_DIR=${0:A:h}
PROJECT_ROOT=${SCRIPT_DIR:h}
COLIMA_ENV="${PROJECT_ROOT}/config/colima.env"
HARBOUR_ENV_DIR="${HOME}/.config/agent-harbour"
HARBOUR_ENV="${HARBOUR_ENV_DIR}/env"
if [[ -f "${COLIMA_ENV}" ]]; then
  source "${COLIMA_ENV}"
fi

if [[ -f "${HARBOUR_ENV}" ]]; then
  source "${HARBOUR_ENV}"
fi

refresh_context_files() {
  RUNTIME_ENV="${HARBOUR_CONTEXT_HOST_PATH:-}/runtime.env"
  REPOS_FILE="${HARBOUR_CONTEXT_HOST_PATH:-}/repos.yaml"
}

load_runtime_env() {
  refresh_context_files
  if [[ -z "${HARBOUR_WORKSPACE_ROOT:-}" && -f "${RUNTIME_ENV}" ]]; then
    source "${RUNTIME_ENV}"
  fi
}

load_runtime_env

require_var() {
  local name=$1
  if [[ -z "${(P)name:-}" ]]; then
    printf "%s is not set. Configure it in %s.\n" "${name}" "${HARBOUR_ENV}" >&2
    exit 1
  fi
}

persist_harbour_env() {
  require_var HARBOUR_CONTEXT_HOST_PATH
  mkdir -p "${HARBOUR_ENV_DIR}"
  cat > "${HARBOUR_ENV}" <<EOF
HARBOUR_CONTEXT_HOST_PATH=${HARBOUR_CONTEXT_HOST_PATH}
HARBOUR_WORKSPACE_ROOT=${HARBOUR_WORKSPACE_ROOT:-}
HARBOUR_ACTIVE_AGENT=${HARBOUR_ACTIVE_AGENT:-}
EOF
  refresh_context_files
}

repo_lines() {
  require_var HARBOUR_CONTEXT_HOST_PATH
  if [[ ! -f "${REPOS_FILE}" ]]; then
    printf "%s is missing. Create it in harbour-context.\n" "${REPOS_FILE}" >&2
    exit 1
  fi
  awk '
    $1 == "-" && $2 == "host_path:" {print $3}
    $1 == "host_path:" {print $2}
  ' "${REPOS_FILE}"
}

desired_mount_lines() {
  require_var HARBOUR_CONTEXT_HOST_PATH
  printf "%s|rw\n" "${HARBOUR_CONTEXT_HOST_PATH}"
  while IFS= read -r host; do
    [[ -n "${host}" ]] || continue
    printf "%s|rw\n" "${host}"
  done < <(repo_lines)
}

current_mount_lines() {
  require_var COLIMA_PROFILE
  local profile_config="${HOME}/.colima/${COLIMA_PROFILE}/colima.yaml"
  if [[ ! -f "${profile_config}" ]]; then
    return 0
  fi

  awk '
    /^mounts:/ {in_mounts=1; next}
    in_mounts && /^[^[:space:]]/ {in_mounts=0}
    in_mounts && $1 == "-" && $2 == "location:" {location=$3}
    in_mounts && $1 == "writable:" {
      mode = ($2 == "true") ? "rw" : "ro"
      printf "%s|%s\n", location, mode
    }
  ' "${profile_config}"
}

state_root() {
  require_var HARBOUR_CONTEXT_HOST_PATH
  printf "%s\n" "${HARBOUR_CONTEXT_HOST_PATH}"
}

bool_flag() {
  local value=$1
  [[ "${value:l}" == "true" ]]
}

colima_status() {
  require_var COLIMA_PROFILE
  colima status -p "${COLIMA_PROFILE}" >/dev/null 2>&1
}
