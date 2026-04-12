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
  mkdir -p "${target_skills_dir}"
  if [[ -d "${harbour_harness_skills_dir}" ]]; then
    while IFS= read -r skill_dir; do
      skill_name=$(basename "${skill_dir}")
      ln -sfn "${skill_dir}" "${target_skills_dir}/${skill_name}"
    done < <(find "${harbour_harness_skills_dir}" -mindepth 1 -maxdepth 1 -type d | sort)
  fi
}

if ! command -v make >/dev/null 2>&1 || ! command -v rg >/dev/null 2>&1 || ! command -v gh >/dev/null 2>&1 || ! command -v file >/dev/null 2>&1; then
  sudo apt-get update
  sudo apt-get install -y file gh make ripgrep
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
    sync_skills "${codex_skills_dir}"

    sudo ln -sfn "${harbour_harness_agents_path}" "${workspace_path}/AGENTS.md"
    sudo rm -f "${workspace_path}/CLAUDE.md"
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
    sync_skills "${claude_skills_dir}"

    sudo ln -sfn "${harbour_harness_agents_path}" "${workspace_path}/CLAUDE.md"
    sudo rm -f "${workspace_path}/AGENTS.md"
    ;;
esac
