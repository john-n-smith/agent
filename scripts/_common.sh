#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR=$(cd -P -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
PROJECT_ROOT=${HARBOUR_PROJECT_ROOT:-$(cd -P -- "${SCRIPT_DIR}/.." && pwd -P)}
HARBOUR_ENV_DIR="${HOME}/.config/harbour"
HARBOUR_ENV="${HARBOUR_ENV_DIR}/env"
HARBOUR_ENV_TEMPLATE="${PROJECT_ROOT}/config/harbour.env.example"

load_harbour_env() {
  if [[ -f "${HARBOUR_ENV}" ]]; then
    source "${HARBOUR_ENV}"
  fi
}

bootstrap_harbour_env() {
  mkdir -p "${HARBOUR_ENV_DIR}"
  if [[ -f "${HARBOUR_ENV}" ]]; then
    return 0
  fi
  if [[ ! -f "${HARBOUR_ENV_TEMPLATE}" ]]; then
    printf "%s is missing.\n" "${HARBOUR_ENV_TEMPLATE}" >&2
    exit 1
  fi
  cp "${HARBOUR_ENV_TEMPLATE}" "${HARBOUR_ENV}"
  printf "Created Harbour env %s from %s.\n" "${HARBOUR_ENV}" "${HARBOUR_ENV_TEMPLATE}"
}

refresh_context_files() {
  REPOS_FILE="${HARBOUR_HARNESS_PATH:-}/repos.yaml"
}

load_harbour_env
refresh_context_files

expand_home_path() {
  local path=$1
  printf "%s\n" "${path/#\~/${HOME}}"
}

prompt_input() {
  local name=$1
  local prompt=$2
  if [[ -t 0 ]]; then
    read -e -r -p "${prompt}" "${name}" 1>&2
  else
    printf "%s" "${prompt}" >&2
    read -r "${name}"
  fi
}

require_var() {
  local name=$1
  if [[ -z "${!name:-}" ]]; then
    printf "%s is not set. Configure it in %s or run harbour provision.\n" "${name}" "${HARBOUR_ENV}" >&2
    exit 1
  fi
}

escape_sed_replacement() {
  printf "%s" "$1" | sed 's/[&|\\]/\\&/g'
}

save_env_var() {
  local name=$1
  local value=$2
  local escaped_value tmp

  escaped_value=$(escape_sed_replacement "${value}")
  tmp=$(mktemp)

  if grep -q "^${name}=" "${HARBOUR_ENV}"; then
    sed -e "s|^${name}=.*$|${name}=${escaped_value}|" "${HARBOUR_ENV}" > "${tmp}"
  else
    cat "${HARBOUR_ENV}" > "${tmp}"
    printf "%s=%s\n" "${name}" "${value}" >> "${tmp}"
  fi

  mv "${tmp}" "${HARBOUR_ENV}"
}

absolute_path() {
  local path=$1
  local parent_dir
  local parent
  path=$(expand_home_path "${path}")
  if [[ -d "${path}" ]]; then
    (cd -P -- "${path}" && pwd -P)
  else
    parent_dir=$(dirname -- "${path}")
    if ! parent=$(cd -P -- "${parent_dir}" >/dev/null 2>&1 && pwd -P); then
      printf "%s\n" "${path}"
      return 0
    fi
    printf "%s/%s\n" "${parent}" "$(basename -- "${path}")"
  fi
}

resolved_repo_lines() {
  local workspace_root=""
  require_var HARBOUR_HARNESS_PATH
  if [[ ! -f "${REPOS_FILE}" ]]; then
    printf "%s is missing. Create it in harbour-harness.\n" "${REPOS_FILE}" >&2
    exit 1
  fi
  if [[ -n "${HARBOUR_WORKSPACE_ROOT:-}" ]]; then
    workspace_root=$(absolute_path "${HARBOUR_WORKSPACE_ROOT}")
  fi
  while IFS= read -r raw_host; do
    [[ -n "${raw_host}" ]] || continue
    raw_host=$(expand_home_path "${raw_host}")
    if [[ "${raw_host}" = /* ]]; then
      printf "%s\n" "${raw_host}"
      continue
    fi

    require_var HARBOUR_WORKSPACE_ROOT
    printf "%s/%s\n" "${workspace_root}" "${raw_host}"
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
  require_var HARBOUR_COLIM_PROFILE
  local profile_config="${HOME}/.colima/${HARBOUR_COLIM_PROFILE}/colima.yaml"
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
  value=$(printf "%s" "${value}" | tr '[:upper:]' '[:lower:]')
  [[ "${value}" == "true" ]]
}

colima_status() {
  require_var HARBOUR_COLIM_PROFILE
  colima status -p "${HARBOUR_COLIM_PROFILE}" >/dev/null 2>&1
}
