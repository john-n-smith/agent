set -euo pipefail

selected_agent=$1
requested_version=$2
harbour_harness_agents_path=$3
harbour_harness_skills_dir=$4
harbour_harness_agents_b64=$5
host_uid=$6
host_gid=$7
workspace_path=$8

agent_bin_dir="${HOME}/.local/bin"
codex_path="${agent_bin_dir}/codex"
claude_path="${agent_bin_dir}/claude"
codex_agents_path="${HOME}/.codex/AGENTS.md"
claude_agents_path="${HOME}/.claude/CLAUDE.md"
codex_skills_dir="${HOME}/.codex/skills"
claude_skills_dir="${HOME}/.claude/skills"
tmpdir=$(mktemp -d)
trap 'rm -rf "${tmpdir}"' EXIT

arch=$(uname -m)
case "${arch}" in
  aarch64|arm64)
    archive_name="codex-aarch64-unknown-linux-musl.tar.gz"
    binary_name="codex-aarch64-unknown-linux-musl"
    ;;
  x86_64|amd64)
    archive_name="codex-x86_64-unknown-linux-musl.tar.gz"
    binary_name="codex-x86_64-unknown-linux-musl"
    ;;
  *)
    echo "Unsupported VM architecture: ${arch}" >&2
    exit 1
    ;;
esac

mkdir -p "${agent_bin_dir}"

sync_skills() {
  local target_skills_dir=$1
  mkdir -p "$(dirname "${target_skills_dir}")"
  rm -rf "${target_skills_dir}"
  if [[ -d "${harbour_harness_skills_dir}" ]]; then
    ln -s "${harbour_harness_skills_dir}" "${target_skills_dir}"
  else
    mkdir -p "${target_skills_dir}"
  fi
}

harness_dir=$(dirname "${harbour_harness_agents_path}")
harness_packages_file="${harness_dir}/apt-packages.txt"

required_packages=(file gh make ripgrep)
harness_packages=()
if [[ -f "${harness_packages_file}" ]]; then
  while IFS= read -r line || [[ -n "${line}" ]]; do
    line="${line%%#*}"
    line="${line#"${line%%[![:space:]]*}"}"
    line="${line%"${line##*[![:space:]]}"}"
    if [[ -n "${line}" ]]; then
      harness_packages+=("${line}")
    fi
  done < "${harness_packages_file}"
fi

all_packages=("${required_packages[@]}" "${harness_packages[@]}")
missing_packages=()
for pkg in "${all_packages[@]}"; do
  if ! dpkg-query -W -f='${Status}' "${pkg}" 2>/dev/null | grep -q '^install ok installed$'; then
    missing_packages+=("${pkg}")
  fi
done

if (( ${#missing_packages[@]} > 0 )); then
  sudo apt-get update
  sudo apt-get install -y "${missing_packages[@]}"
fi

harness_provision_hook="${harness_dir}/provision.sh"
if [[ -f "${harness_provision_hook}" ]]; then
  echo "Running harness provision hook: ${harness_provision_hook}"
  hook_status=0
  bash "${harness_provision_hook}" || hook_status=$?
  if (( hook_status != 0 )); then
    echo "WARNING: harness provision hook exited ${hook_status}: ${harness_provision_hook}" >&2
    echo "WARNING: continuing provisioning without it; re-run 'harbour provision' to retry." >&2
  fi
fi

if [[ ! -f "${harbour_harness_agents_path}" ]]; then
  tmp_agents=$(mktemp)
  trap 'rm -rf "${tmpdir}" "${tmp_agents}"' EXIT
  printf '%s' "${harbour_harness_agents_b64}" | base64 -d > "${tmp_agents}"
  sudo install -o "${host_uid}" -g "${host_gid}" -m 0644 "${tmp_agents}" "${harbour_harness_agents_path}"
fi

case "${selected_agent}" in
  codex)
    current_version=""
    if [[ -x "${codex_path}" ]]; then
      current_version=$("${codex_path}" --version 2>/dev/null | awk '{print $2}')
    fi

    version="${requested_version}"
    if [[ "${version}" == "latest" ]]; then
      latest_url=$(curl -fsSL -o /dev/null -w '%{url_effective}' https://github.com/openai/codex/releases/latest)
      version=${latest_url##*/}
      version=${version#rust-v}
      if [[ -z "${version}" || "${version}" == "latest" ]]; then
        echo "Failed to resolve the latest Codex release version" >&2
        exit 1
      fi
    fi

    url="https://github.com/openai/codex/releases/download/rust-v${version}/${archive_name}"

    echo "Target Codex version: ${version}"
    if [[ "${current_version}" != "${version}" ]]; then
      curl -fsSL "${url}" -o "${tmpdir}/codex.tar.gz"
      tar -xzf "${tmpdir}/codex.tar.gz" -C "${tmpdir}"
      install -m 0755 "${tmpdir}/${binary_name}" "${codex_path}"
    fi

    rm -f "${claude_path}"
    mkdir -p "$(dirname "${codex_agents_path}")"
    ln -sfn "${harbour_harness_agents_path}" "${codex_agents_path}"
    sync_skills "${codex_skills_dir}"
    ;;
  claude)
    current_version=""
    if [[ -x "${claude_path}" ]]; then
      current_version=$("${claude_path}" --version 2>/dev/null | awk '{print $NF}' | sed 's/^v//')
    fi

    version="${requested_version}"
    echo "Target Claude Code version: ${version}"
    if [[ "${version}" == "latest" ]]; then
      curl -fsSL https://claude.ai/install.sh | bash
      version=$("${claude_path}" --version 2>/dev/null | awk '{print $NF}' | sed 's/^v//')
      if [[ -z "${version}" ]]; then
        echo "Failed to detect the installed Claude Code version" >&2
        exit 1
      fi
    elif [[ "${current_version}" != "${version}" ]]; then
      curl -fsSL https://claude.ai/install.sh | bash -s "${version}"
    fi

    rm -f "${codex_path}"
    mkdir -p "$(dirname "${claude_agents_path}")"
    ln -sfn "${harbour_harness_agents_path}" "${claude_agents_path}"
    sync_skills "${claude_skills_dir}"
    ;;
esac
