#!/bin/sh
# Install a released AgentMesh CLI and its matching agent skill.
set -eu

fail() { printf 'Error: %s\n' "$*" >&2; exit 1; }
fetch() { curl --fail --silent --show-error --location --retry 2 --connect-timeout 15 --max-time 180 "$1" -o "$2"; }

command -v curl >/dev/null 2>&1 || fail 'curl is required.'
case "$(uname -s)" in
    Linux) os=linux ;;
    Darwin) os=darwin ;;
    *) fail 'Unsupported OS. Use install.ps1 on Windows.' ;;
esac
case "$(uname -m)" in
    x86_64|amd64) arch=amd64 ;;
    arm64|aarch64) arch=arm64 ;;
    *) fail 'Unsupported architecture. Supported: amd64 and arm64.' ;;
esac
if command -v sha256sum >/dev/null 2>&1; then
    hash_file() { sha256sum "$1" | awk '{print $1}'; }
elif command -v shasum >/dev/null 2>&1; then
    hash_file() { shasum -a 256 "$1" | awk '{print $1}'; }
else
    fail 'sha256sum or shasum is required.'
fi

repo=https://github.com/alanjm0532/agentmesh
version=${AGENTMESH_VERSION:-latest}
if [ "$version" = latest ]; then
    release_url=$(curl --fail --silent --show-error --head --location --retry 2 \
        --connect-timeout 15 --max-time 60 --output /dev/null --write-out '%{url_effective}' "$repo/releases/latest")
    case "$release_url" in
        "$repo"/releases/tag/*) version=${release_url##*/} ;;
        *) fail 'Cannot resolve the latest published release.' ;;
    esac
fi
case "$version" in ''|*[!a-zA-Z0-9._-]*) fail 'Release tags must contain only letters, digits, dots, underscores, and hyphens.' ;; esac
base_url=$repo/releases/download/$version
binary=agentmesh_${os}_${arch}
stage=$(mktemp -d)
trap 'rm -rf "$stage"' EXIT
trap 'exit 1' HUP INT TERM

printf 'Installing AgentMesh %s for %s/%s...\n' "$version" "$os" "$arch"
fetch "$base_url/$binary" "$stage/$binary"
fetch "$base_url/agentmesh-SKILL.md" "$stage/agentmesh-SKILL.md"
fetch "$base_url/SHA256SUMS" "$stage/SHA256SUMS"
for asset in "$binary" agentmesh-SKILL.md; do
    expected=$(awk -v name="$asset" '$2 == name {print $1}' "$stage/SHA256SUMS")
    [ "${#expected}" -eq 64 ] || fail "Missing or invalid checksum for $asset. This release must include installer assets."
    actual=$(hash_file "$stage/$asset")
    [ "$expected" = "$actual" ] || fail "Checksum mismatch for $asset."
done

install_dir=${AGENTMESH_INSTALL_DIR:-$HOME/.local/bin}
case "$install_dir" in /*) ;; *) fail 'AGENTMESH_INSTALL_DIR must be an absolute path.' ;; esac
mkdir -p "$install_dir"
# Rename within the destination directory so upgrades can replace a running CLI.
binary_tmp=$(mktemp "$install_dir/.agentmesh.XXXXXX")
if ! (cp "$stage/$binary" "$binary_tmp" && chmod 755 "$binary_tmp" && mv -f "$binary_tmp" "$install_dir/agentmesh"); then
    rm -f "$binary_tmp"
    fail "Cannot install the binary in $install_dir."
fi

skill_dir=$HOME/.agents/skills/agentmesh
mkdir -p "$skill_dir"
cp "$stage/agentmesh-SKILL.md" "$skill_dir/SKILL.md"
printf 'Installed skill: %s/SKILL.md\n' "$skill_dir"

# Persist PATH for new terminals. The parent shell cannot be changed by a script.
quoted_dir=$(printf '%s' "$install_dir" | sed "s/'/'\\\\''/g")
path_line="export PATH='$quoted_dir':\"\$PATH\""
add_path() {
    if ! grep -Fqx "$path_line" "$1" 2>/dev/null; then
        printf '\n# AgentMesh CLI\n%s\n' "$path_line" >> "$1"
    fi
}
case "${SHELL:-/bin/sh}" in
    */bash)
        add_path "$HOME/.bashrc"
        add_path "$HOME/.profile"
        if [ -f "$HOME/.bash_profile" ]; then add_path "$HOME/.bash_profile"; fi
        ;;
    */zsh) add_path "${ZDOTDIR:-$HOME}/.zshrc" ;;
    */sh|*/dash|*/ksh) add_path "$HOME/.profile" ;;
    *) printf 'Add %s to your shell PATH.\n' "$install_dir" ;;
esac
"$install_dir/agentmesh" --help >/dev/null
printf '\nInstalled: %s/agentmesh\n' "$install_dir"
printf 'For this terminal, run: %s\n' "$path_line"
printf 'Start a new agent session to discover the skill.\n'
