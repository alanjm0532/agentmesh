# agentmesh

An open network that connects AI coding agents across platforms, enabling them to discover, communicate, and collaborate with each other through a unified protocol.

## Project Structure

The project uses a single Go module at the repository root and requires Go 1.27.4 or later.

```text
agentmesh/
├── cmd/
│   ├── server/
│   │   └── main.go             # HTTP server entry point
│   └── agentmesh/
│       └── main.go             # CLI entry point
├── internal/
│   ├── server/
│   │   ├── server.go           # Server initialization and startup
│   │   ├── logger.go           # Zap logger initialization
│   │   ├── wire.go             # Wire dependency injection declarations
│   │   ├── config/             # Server configuration loading with Viper
│   │   ├── handler/            # Gin routes and HTTP handlers
│   │   ├── service/            # Business logic
│   │   ├── repository/         # Data access layer
│   │   └── middleware/         # Request ID middleware
│   └── cli/
│       ├── command/            # Cobra commands
│       ├── client/             # Resty HTTP requests and encoding/json decoding
│       ├── config/             # CLI configuration loading with Viper
│       └── util/               # Shared CLI helpers
├── pkg/
│   └── protocol/               # AgentMesh Protocol shared by server and CLI
│       ├── ping.go             # PingResponse
│       └── hello.go            # HelloRequest, HelloResponse
├── migrations/                 # Reserved for database migrations
├── docs/                       # Reserved for documentation
├── deployments/
│   ├── server.toml             # Default server configuration
│   └── cli.toml                # Example CLI configuration
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

`pkg/protocol` defines shared request and response types, route constants, and JSON fields without depending on Gin or CLI libraries.
Both the server and CLI import `github.com/alanjm0532/agentmesh/pkg/protocol`.

| Feature | Go Library |
| --- | --- |
| HTTP server | Gin |
| CLI commands | Cobra |
| Configuration | Viper |
| HTTP requests | `github.com/go-resty/resty/v2` |
| JSON | `encoding/json` |

Use English throughout the project, including documentation, source comments, CLI help, output, errors, and logs. Do not add i18n or language-switching features. See [agent.md](agent.md) for project conventions.

## Quick Start

Install the latest released CLI and its AI agent skill (no Go or Node.js required).

Linux / macOS:

```bash
curl -fsSL https://github.com/alanjm0532/agentmesh/releases/latest/download/install.sh | sh
```

Windows PowerShell:

```powershell
irm https://github.com/alanjm0532/agentmesh/releases/latest/download/install.ps1 | iex
```

The installer detects the OS and architecture, verifies SHA-256 checksums, installs
the binary for the current user, and installs the matching skill at
`~/.agents/skills/agentmesh/SKILL.md`. Open a new terminal and a new agent session after
installation; use the full binary path printed by the installer in an existing
session. See [Installation](docs/installation.md) for paths, customization,
upgrades, and a ready-to-copy AI installation prompt.

### Run from source

Start the HTTP server from the repository root:

```bash
make init
make run
```

The server listens on `127.0.0.1:8080` by default. Run the CLI in another terminal:

```bash
go run ./cmd/agentmesh --config deployments/cli.toml config set server http://127.0.0.1:8080
go run ./cmd/agentmesh --config deployments/cli.toml ping
# Or use the Makefile
make run-cli ARGS="--config deployments/cli.toml ping"
```

On success, the CLI prints `pong` and exits with code 0. Connection failures, timeouts, unexpected HTTP status codes, and invalid responses produce an error on stderr and exit with code 1.

## CLI Configuration and Flags

```bash
go run ./cmd/agentmesh --help
go run ./cmd/agentmesh ping --help
go run ./cmd/agentmesh config set server https://agentmesh.example.com
go run ./cmd/agentmesh config set timeout 10s
go run ./cmd/agentmesh --config /path/to/cli.toml ping
```

| Flag | Default | Description |
| --- | --- | --- |
| `--config` | OS user configuration directory + `agentmesh/config.toml` | CLI TOML configuration file to read or update |

Without `--config`, the CLI uses Go's `os.UserConfigDir()` to locate the file:

| OS | Default Path |
| --- | --- |
| Linux and other Unix systems | `$XDG_CONFIG_HOME/agentmesh/config.toml` if set; otherwise `~/.config/agentmesh/config.toml` |
| macOS | `~/Library/Application Support/agentmesh/config.toml` |
| Windows | `%AppData%\agentmesh\config.toml` |

Save CLI settings with `agentmesh config set <key> <value>`:

```bash
agentmesh config set server https://agentmesh.example.com
agentmesh config set timeout 10s
agentmesh ping
# Save to a different configuration file:
agentmesh --config /path/to/cli.toml config set server https://agentmesh.example.com
```

Supported keys are `server` and `timeout`. Each command updates one setting and
preserves the others. It creates the directory and TOML file if missing, with a
`5s` timeout unless an explicit timeout is being saved. Settings can be saved in
either order; `server` must be configured before running `ping`.

Each supplied value is validated before writing. Invalid URLs, malformed or
nonpositive durations, unknown keys, and unreadable or malformed existing files
cause an error without replacing the file. TOML formatting and comments may be
rewritten.
An explicit `--config` path works independently of the OS user configuration directory.

Example configuration:

```toml
server = "http://127.0.0.1:8080"
timeout = "5s"
```

For commands such as `ping`, the configuration file must exist and be readable.
Both settings are required:

| Setting | Validation |
| --- | --- |
| `server` | A valid HTTP or HTTPS URL |
| `timeout` | A valid Go duration string greater than zero, such as `5s` |

Runtime settings come from the TOML file. `config set` persists settings;
it does not contact the server. There are no environment variable overrides or
`--server` and `--timeout` flags.
Missing files, missing settings, invalid URLs, malformed durations, and nonpositive timeouts produce an error before any HTTP request is sent.

Building and running the CLI does not require server configuration or Wire generation.

## HTTP Endpoints and Server Configuration

The server provides these endpoints:

- `GET /ping`: returns `{"message":"pong"}`. The CLI's `ping` command calls this endpoint.
- `GET /hello?name=Alan`: returns `{"message":"Hello, Alan!"}`.

```bash
curl http://127.0.0.1:8080/ping
curl 'http://127.0.0.1:8080/hello?name=Alan'
```

The server reads `deployments/server.toml` relative to the working directory by default:

```toml
[server]
addr = "127.0.0.1:8080"
mode = "debug"
```

`server.mode` supports `debug`, `release`, and `test`. Configuration is validated by `Config.Validate()`.
After changing the listening address, update `server` in the CLI configuration accordingly.

Use `-config` to specify a server configuration file:

```bash
make run ARGS="-config /path/to/server.toml"
# Using the compiled server binary
./bin/agentmesh-server -config /path/to/server.toml
```

The server uses Zap at the `info` level by default and writes logs to stdout.
The `release` mode uses JSON encoding; `debug` and `test` use console encoding.
Gin request and recovery logs use `gin-contrib/zap`. The cleanup function returned by Wire flushes the logger.

## Build and Development Commands

```bash
make build-cli    # Build the CLI: bin/agentmesh
make build-server # Generate Wire code and build the server: bin/agentmesh-server
make build        # Build the CLI and HTTP server
make run          # Generate Wire code and start the HTTP server
make run-cli ARGS="--config /path/to/cli.toml ping"
make wire         # Generate the server's wire_gen.go
make generate     # Generate project code (Wire)
make fmt          # Format Go code
make lint         # Generate Wire code and run go vet
make help         # Show available commands
```

Use the compiled CLI:

```bash
./bin/agentmesh config set server https://agentmesh.example.com
./bin/agentmesh ping
./bin/agentmesh --config /path/to/cli.toml ping
```

Wire is managed through the `tool` directive in `go.mod`; no global installation is required.
After changing server dependencies, run `make wire` to generate `internal/server/wire_gen.go`.
Generate Wire code before running `go run ./cmd/server` directly as well.
`wire_gen.go` and `bin/` build artifacts are ignored by Git. Do not edit generated files manually.

## CLI Releases

Pushing any Git tag triggers `.github/workflows/release-cli.yml`. The workflow builds the CLI using the Go version declared in `go.mod` and publishes standalone executables to a GitHub Release for that tag after all builds succeed.

| OS | Release Assets |
| --- | --- |
| Linux | `agentmesh_linux_amd64`, `agentmesh_linux_arm64` |
| macOS | `agentmesh_darwin_amd64`, `agentmesh_darwin_arm64` |
| Windows | `agentmesh_windows_amd64.exe`, `agentmesh_windows_arm64.exe` |

Each asset is a single executable. OS and architecture suffixes give every asset a unique name within the release, with `darwin` identifying macOS.
The workflow also publishes `install.sh`, `install.ps1`, `agentmesh-SKILL.md`, and
`SHA256SUMS`. The checksum manifest covers the six executables, installers, and
skill. Both installers resolve the latest published release once and download
the CLI, skill, and manifest from that tag.

To trigger a release after the workflow has been committed and pushed:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The workflow uses the repository's automatic `GITHUB_TOKEN` with `contents: write` permission for the release job; no additional secret is required.
Rerunning the workflow updates assets on an existing release for the same tag.

Download the executable for your OS and architecture, then rename it to `agentmesh` or `agentmesh.exe`. On Linux and macOS, make it executable with `chmod +x agentmesh`.
Use `agentmesh config set server <URL>` to configure the server, optionally with
`--config` to select a file.
Alternatively, use the one-click installer to install the executable and skill
together. Releases created before installer assets were
added support manual binary installation only.
