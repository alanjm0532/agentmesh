#!/bin/sh
# Install a released AgentMesh CLI and its matching agent skill.
set -eu

fail() { printf 'Error: %s\n' "$*" >&2; exit 1; }
fetch() {
    fetch_url=$1
    fetch_path=$2
    shift 2
    curl --fail --silent --show-error --location --connect-timeout 15 --max-time 45 "$@" "$fetch_url" -o "$fetch_path"
}

# Read only the release tag and top-level asset names/IDs. Tokenize JSON rather
# than relying on GitHub whitespace or field order; no jq/Python dependency.
release_info() {
    awk '
    {
        while (match($0, /"([^"\\]|\\.)*"|[][{}:,]|[0-9]+/)) {
            token = substr($0, RSTART, RLENGTH)
            $0 = substr($0, RSTART + RLENGTH)
            value = token
            if (substr(token, 1, 1) == "\"") value = substr(token, 2, length(token) - 2)
            if (token == "{" || token == "[") {
                depth++
                if (token == "[" && depth == 2 && key == "assets") assets = depth
                if (token == "{" && assets && depth == assets + 1) { name = ""; id = "" }
            } else if (token == "}" || token == "]") {
                if (token == "}" && assets && depth == assets + 1 && name ~ /^[a-zA-Z0-9._-]+$/ && id ~ /^[1-9][0-9]*$/)
                    print "asset", name, id
                if (token == "]" && depth == assets) assets = 0
                depth--
            } else if (token == ":") {
                key = previous_value
            } else if (previous == ":") {
                if (depth == 1 && key == "tag_name" && value ~ /^[a-zA-Z0-9][a-zA-Z0-9._-]*$/) print "tag", value
                if (assets && depth == assets + 1) {
                    if (key == "name") name = value
                    if (key == "id") id = value
                }
            }
            previous = token
            previous_value = value
        }
    }' "$1"
}

fetch_release_info() {
    fetch "$1" "$stage/release.json" -H 'Accept: application/vnd.github+json' || return 1
    release_info "$stage/release.json" > "$stage/release-info"
}

fetch_asset() {
    if fetch "$base_url/$1" "$stage/$1"; then return 0; fi
    printf 'Direct download failed for %s; trying the GitHub API...\n' "$1" >&2
    if [ ! -f "$stage/release-info" ]; then
        fetch_release_info "$releases_api/tags/$version" || fail 'GitHub API release lookup failed after direct download failed.'
    fi
    api_tag=$(awk '$1 == "tag" {print $2}' "$stage/release-info")
    [ "$api_tag" = "$version" ] || fail 'GitHub API release tag mismatch.'
    asset_id=$(awk -v name="$1" '$1 == "asset" && $2 == name {print $3}' "$stage/release-info")
    case "$asset_id" in ''|*[!0-9]*) fail "Missing or ambiguous GitHub API asset: $1." ;; esac
    fetch "$releases_api/assets/$asset_id" "$stage/$1" -H 'Accept: application/octet-stream' \
        || fail "Both direct and GitHub API downloads failed for $1."
}

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
releases_api=https://api.github.com/repos/alanjm0532/agentmesh/releases
stage=$(mktemp -d)
trap 'rm -rf "$stage"' EXIT
trap 'exit 1' HUP INT TERM
version=${AGENTMESH_VERSION:-latest}
if [ "$version" = latest ]; then
    release_url=$(curl --fail --silent --show-error --head --location \
        --connect-timeout 15 --max-time 45 --output /dev/null --write-out '%{url_effective}' "$repo/releases/latest") || release_url=''
    case "$release_url" in
        "$repo"/releases/tag/*) version=${release_url##*/} ;;
        *)
            printf 'Cannot resolve latest release directly; trying the GitHub API...\n' >&2
            fetch_release_info "$releases_api/latest" || fail 'Cannot resolve the latest published release.'
            version=$(awk '$1 == "tag" {print $2}' "$stage/release-info")
            ;;
    esac
fi
case "$version" in ''|*[!a-zA-Z0-9._-]*) fail 'Release tags must contain only letters, digits, dots, underscores, and hyphens.' ;; esac
[ "$version" != latest ] || fail 'Cannot resolve the latest published release.'
base_url=$repo/releases/download/$version
binary=agentmesh_${os}_${arch}

printf 'Installing AgentMesh %s for %s/%s...\n' "$version" "$os" "$arch"
fetch_asset "$binary"
fetch_asset SHA256SUMS
expected=$(awk -v name="$binary" '$2 == name {print $1}' "$stage/SHA256SUMS")
[ "${#expected}" -eq 64 ] || fail "Missing or invalid checksum for $binary."
actual=$(hash_file "$stage/$binary")
[ "$expected" = "$actual" ] || fail "Checksum mismatch for $binary."

install_dir=${AGENTMESH_INSTALL_DIR:-$HOME/.local/bin}
case "$install_dir" in /*) ;; *) fail 'AGENTMESH_INSTALL_DIR must be an absolute path.' ;; esac
mkdir -p "$install_dir"
# Rename within the destination directory so upgrades can replace a running CLI.
binary_tmp=$(mktemp "$install_dir/.agentmesh.XXXXXX")
if ! (cp "$stage/$binary" "$binary_tmp" && chmod 755 "$binary_tmp" && mv -f "$binary_tmp" "$install_dir/agentmesh"); then
    rm -f "$binary_tmp"
    fail "Cannot install the binary in $install_dir."
fi

"$install_dir/agentmesh" skill install --version "$version"

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
