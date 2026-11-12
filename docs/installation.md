# Installation

## Run the installer

Linux / macOS:

```bash
curl -fsSL https://github.com/alanjm0532/agentmesh/releases/latest/download/install.sh | sh
```

Windows PowerShell:

```powershell
irm https://github.com/alanjm0532/agentmesh/releases/latest/download/install.ps1 | iex
```

The script downloads the CLI, configures PATH, and runs `agentmesh skill install`
to install the skill at `~/.agents/skills/agentmesh/`.

## Verify installation

Open a new terminal, then run:

```bash
agentmesh --version
```

Start a new agent session to discover the skill.
