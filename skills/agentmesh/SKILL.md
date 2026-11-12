---
name: agentmesh
description: Use the AgentMesh CLI to configure an AgentMesh server connection and check its availability. Use when a user asks to use agentmesh or diagnose an AgentMesh connection.
---

# AgentMesh CLI

Use the installed `agentmesh` executable. Start with `agentmesh --help` and
`agentmesh ping --help` to check the commands supported by the installed release.
The current CLI supports `config set` and `ping`; discovery, messaging, authentication, and
collaboration commands are not implemented yet.

## Configuration

Save settings with the CLI; change timeout only when requested:

```bash
agentmesh config set server https://agentmesh.example.com
agentmesh config set timeout 10s
```

Replace the example URL with the user's server address. If the intended server
is unknown, ask for its URL before changing the configuration. Use
`agentmesh config set <key> <value>` with `server` or `timeout`. Each command
validates the supplied value and preserves other settings. It creates the
configuration if missing, with `timeout = "5s"` unless setting timeout explicitly.
Either setting can be saved first; configure server before running `ping`.
These commands do not contact the server. Prefer them over manually editing TOML.

`ping` requires the saved configuration. `server` must be an HTTP or HTTPS URL and
`timeout` a positive Go duration.

Default file locations:

| System | Path |
| --- | --- |
| Linux | `$XDG_CONFIG_HOME/agentmesh/config.toml`, or `~/.config/agentmesh/config.toml` when unset |
| macOS | `~/Library/Application Support/agentmesh/config.toml` |
| Windows | `%AppData%\agentmesh\config.toml` |

Select another file with `agentmesh --config /path/to/config.toml ping`.
Use `agentmesh --config /path/to/config.toml config set server <URL>` to update
that file instead of the default.
There are no `--server` or `--timeout` flags or environment overrides for these
runtime settings.

## Check connectivity

```bash
agentmesh ping
```

This sends `GET /ping` to the configured server. Success prints `pong` and exits
with code 0. Errors go to stderr with exit code 1. Read the exit status and stderr;
do not assume success from a running process alone.

For a missing or invalid configuration, inspect the selected file and use
`config set` to correct the reported setting. For connection failures, confirm the configured URL and
whether that server is running. Installing the CLI does not start a server.
Do not invent login, JSON-output, or server-start commands.

If the shell cannot find `agentmesh`, check the installer-reported binary path
(`~/.local/bin/agentmesh` by default on Linux/macOS, or
`%LOCALAPPDATA%\Programs\agentmesh\agentmesh.exe` on Windows) and use its absolute
path or refresh the shell's PATH. Newly installed skills may require a new agent
session before they are discovered.
