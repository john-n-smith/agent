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
  RUNTIME_ENV="${HARBOUR_HARNESS_PATH:-}/runtime.env"
  REPOS_FILE="${HARBOUR_HARNESS_PATH:-}/repos.yaml"
}

load_runtime_env() {
  refresh_context_files
  if [[ -z "${HARBOUR_WORKSPACE_ROOT:-}" && -f "${RUNTIME_ENV}" ]]; then
    source "${RUNTIME_ENV}"
  fi
}

load_runtime_env

expand_home_path() {
  local path=$1
  printf "%s\n" "${path/#\~/${HOME}}"
}

require_var() {
  local name=$1
  if [[ -z "${(P)name:-}" ]]; then
    printf "%s is not set. Configure it in %s.\n" "${name}" "${HARBOUR_ENV}" >&2
    exit 1
  fi
}

persist_harbour_env() {
  require_var HARBOUR_HARNESS_PATH
  mkdir -p "${HARBOUR_ENV_DIR}"
  cat > "${HARBOUR_ENV}" <<EOF
HARBOUR_HARNESS_PATH=${HARBOUR_HARNESS_PATH}
HARBOUR_WORKSPACE_ROOT=${HARBOUR_WORKSPACE_ROOT:-}
HARBOUR_ACTIVE_AGENT=${HARBOUR_ACTIVE_AGENT:-}
EOF
  refresh_context_files
}

resolved_repo_lines() {
  require_var HARBOUR_HARNESS_PATH
  if [[ ! -f "${REPOS_FILE}" ]]; then
    printf "%s is missing. Create it in harbour-harness.\n" "${REPOS_FILE}" >&2
    exit 1
  fi
  while IFS= read -r raw_host; do
    [[ -n "${raw_host}" ]] || continue
    raw_host=$(expand_home_path "${raw_host}")
    if [[ "${raw_host}" = /* ]]; then
      printf "%s\n" "${raw_host}"
      continue
    fi

    require_var HARBOUR_WORKSPACE_ROOT
    printf "%s/%s\n" "${HARBOUR_WORKSPACE_ROOT:A}" "${raw_host}"
  done < <(
    awk '
      /^[[:space:]]*-[[:space:]]*host_path:[[:space:]]*/ {
        sub(/^[[:space:]]*-[[:space:]]*host_path:[[:space:]]*/, "", $0)
        sub(/[[:space:]]+#.*$/, "", $0)
        print $0
        next
      }
      /^[[:space:]]*host_path:[[:space:]]*/ {
        sub(/^[[:space:]]*host_path:[[:space:]]*/, "", $0)
        sub(/[[:space:]]+#.*$/, "", $0)
        print $0
      }
    ' "${REPOS_FILE}"
  )
}

repo_lines() {
  local warn_missing=${1:-false}
  while IFS= read -r host; do
    [[ -n "${host}" ]] || continue
    if [[ -d "${host}" ]]; then
      printf "%s\n" "${host}"
      continue
    fi

    if bool_flag "${warn_missing}"; then
      printf "Warning: skipping missing repo mount %s\n" "${host}" >&2
    fi
  done < <(resolved_repo_lines)
}

desired_mount_lines() {
  require_var HARBOUR_HARNESS_PATH
  printf "%s|rw\n" "${HARBOUR_HARNESS_PATH}"
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
  require_var HARBOUR_HARNESS_PATH
  printf "%s\n" "${HARBOUR_HARNESS_PATH}"
}

bool_flag() {
  local value=$1
  [[ "${value:l}" == "true" ]]
}

colima_status() {
  require_var COLIMA_PROFILE
  colima status -p "${COLIMA_PROFILE}" >/dev/null 2>&1
}
