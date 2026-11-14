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

The installers and `agentmesh skill install` try the normal GitHub Release
download URL first. If a download fails or times out (45 seconds per attempt),
they automatically retry that file through the official GitHub Releases API.
Both paths download the same release and require SHA-256 verification before
installation. If both fail, the command reports an error.

## Verify installation

Open a new terminal, then run:

```bash
agentmesh --version
```

## Configure the server

Before logging in, set the AgentMesh server URL provided by the user. If the URL
is unknown, ask the user for it before continuing. Replace the example below
with the user's complete HTTP or HTTPS URL:

```bash
agentmesh config set server "https://your-agentmesh-server.example.com"
agentmesh ping
```

A successful connection prints `pong`. Resolve any connection errors before
starting login.

## Log in

Start login. The CLI will try to open the authorization page in the default
browser:

```bash
agentmesh auth login
```

Show the printed authorization code to the user. Ask the user to sign in if
prompted, check that the code in the browser matches, and click **Authorize
AgentMesh CLI**. If the browser does not open automatically, give the user the
printed authorization URL to open manually.

For a remote terminal or an environment without a browser, use
`agentmesh auth login --no-browser` and give the user the printed authorization URL
and code to complete authorization in their browser.

Keep the login command running while the user completes authorization. The CLI
saves the token automatically and prints `Logged in as <username>` when login
succeeds. If the login expires, run the command again and give the user the new
URL and code.

## Verify login

```bash
agentmesh auth status
```

Confirm that the command succeeds and prints `Logged in as <username>` before
reporting setup as complete. Start a new agent session to discover the installed
skill.
