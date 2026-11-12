# Installation

## One command

Linux and macOS (amd64 / arm64):

```bash
curl -fsSL https://github.com/alanjm0532/agentmesh/releases/latest/download/install.sh | sh
```

Windows PowerShell (amd64 / arm64):

```powershell
irm https://github.com/alanjm0532/agentmesh/releases/latest/download/install.ps1 | iex
```

The commands become available after a release containing the installers,
`agentmesh-SKILL.md`, and `SHA256SUMS` has been published. Older releases containing
only executables cannot be installed with these scripts.

To inspect a script before running it, download it instead of piping it to a
shell. For example:

```bash
curl -fL https://github.com/alanjm0532/agentmesh/releases/latest/download/install.sh -o install.sh
sh install.sh
```

Installation requires HTTPS access to GitHub and its release download hosts.
Linux/macOS require `curl` and either `sha256sum` or `shasum`. Windows requires
PowerShell 5.1 or later. No administrator access, Go, npm, or Node.js is required.

## What gets installed

1. Resolve the latest published GitHub Release, excluding drafts and prereleases.
2. Download the standalone executable for the OS and CPU, its matching skill,
   and `SHA256SUMS` from the same tag.
3. Verify the binary and skill checksums before installing either file.
4. Install the binary and skill for the current user.
5. Add the binary directory to PATH and check `agentmesh --help`.

| Item | Linux / macOS | Windows |
| --- | --- | --- |
| Binary | `~/.local/bin/agentmesh` | `%LOCALAPPDATA%\Programs\agentmesh\agentmesh.exe` |
| Shared skill | `~/.agents/skills/agentmesh/SKILL.md` | `%USERPROFILE%\.agents\skills\agentmesh\SKILL.md` |

The skill is installed only in the shared `~/.agents/skills` directory, without a
skill manager or symlink support. Existing `agentmesh/SKILL.md` files are updated;
unrelated skills are preserved. Agents that support this shared directory can
discover the skill. The installer does not manage agent-specific directories.
The installer source files are in `scripts/install.sh` and `scripts/install.ps1`.

On Linux/macOS, the installer adds the binary directory to the selected shell's
startup files: `.bashrc` and `.profile` for Bash (also `.bash_profile` if present),
`${ZDOTDIR:-~}/.zshrc` for Zsh, or `.profile` for POSIX shells. Other shells receive
manual PATH instructions. The installer prints an `export PATH=...` command for
the current terminal. On Windows it updates the user PATH and the PowerShell
process running the script. Other existing processes keep their old environment.

Start a new terminal and a new agent session so the binary and skill are
discovered. An AI performing installation in an existing session can read the
installed `SKILL.md` directly and invoke the printed absolute binary path.

## Verify installation

```bash
agentmesh --help
```

The installer verifies CLI startup with `--help`. Server configuration is handled
when using commands that need a connection:

```bash
agentmesh config set server https://agentmesh.example.com
agentmesh config set timeout 10s
agentmesh ping
```

Replace the example URL with your server address. `config set <key> <value>`
supports `server` and `timeout`, creates the file if missing, and preserves other
settings when updating it. See
[CLI configuration](../README.md#cli-configuration-and-flags).

## Upgrade or pin a release

Rerun the one-command installer to upgrade the binary and skill together. Existing
configuration is retained and identical PATH entries are not appended again.

To choose a release or custom locations, download the script first:

```bash
AGENTMESH_VERSION=v0.1.0 \
AGENTMESH_INSTALL_DIR="$HOME/tools/bin" \
sh install.sh
```

```powershell
Invoke-WebRequest -UseBasicParsing https://github.com/alanjm0532/agentmesh/releases/latest/download/install.ps1 -OutFile install.ps1
.\install.ps1 -Version v0.1.0 -InstallDir "$HOME\tools\bin"
```

`AGENTMESH_VERSION` and `AGENTMESH_INSTALL_DIR` are also
recognized by the PowerShell installer. A pinned release must contain the installer
assets. Supported tag characters are letters, digits, dots, underscores, and
hyphens. The manifest detects damaged downloads; it is fetched from the same
HTTPS release and is not a separately signed authenticity proof.

## Prompt for an AI agent

Copy this prompt into your coding agent:

```text
Install the AgentMesh CLI and its agent skill using this guide:
https://github.com/alanjm0532/agentmesh/blob/main/docs/installation.md

Detect my operating system and CPU architecture. Download and run the official
installer for my system from the latest GitHub Release of alanjm0532/agentmesh.
It installs the matching binary and skill together and verifies their checksums.
Read the installed agentmesh/SKILL.md so you
can use the CLI in this session, and verify installation with agentmesh --help
(use the installer's absolute binary path if PATH has not refreshed).
Report the installation paths and explain if I need a new session for skill discovery.
```
