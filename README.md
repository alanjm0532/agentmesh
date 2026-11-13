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
├── Dockerfile                  # Multi-stage server image build
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
`~/.agents/skills/agentmesh/`, including `SKILL.md` and `references/`.
Open a new terminal and a new agent session after
installation; use the full binary path printed by the installer in an existing
session. See [Installation](docs/installation.md) for installation and verification.

If the CLI is already installed, install or update its skill with:

```bash
agentmesh skill install
agentmesh skill install --version v0.1.0
agentmesh skill install --version latest
agentmesh --version
```

The default skill version matches the CLI release tag. Download failures return
an error without selecting another version.
Local builds default to `latest`. Skill installation works
without server configuration.

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

### Run the server with Docker

Prepare an external server TOML configuration with `server.addr` set to
`0.0.0.0:8080` and `server.mode` set to `release`, then build and start the server
from the repository root:

```bash
docker build -t agentmesh-server .
docker run --rm --name agentmesh-server -p 8080:8080 \
  -v "/path/to/server.toml:/app/deployments/server.toml:ro" \
  agentmesh-server
```

The Dockerfile generates Wire code and builds the server with Go 1.27.4.
The runtime image runs as a non-root user and reads the configuration mounted at
`/app/deployments/server.toml` at startup. Configuration files are excluded from
the build context.

Replace `/path/to/server.toml` with the absolute path to your configuration file.
Check the running server with `curl http://127.0.0.1:8080/ping`.

### Server image releases

Pushing a Git tag triggers `.github/workflows/release-server.yml`, which runs
lint first. After lint passes, it builds the Linux amd64 server image and pushes
it to GitHub Container Registry (GHCR).
The image is named `ghcr.io/<owner>/<repository>-server` and uses the Git tag as
its version. Branch pushes do not publish images, and no `latest` tag is generated.
The workflow uses `GITHUB_TOKEN` with `packages: write` to authenticate.

For `alanjm0532/agentmesh`, pushing `v0.1.0` publishes
`ghcr.io/alanjm0532/agentmesh-server:v0.1.0`:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Run a published version with your configuration mounted:

```bash
docker run --rm --name agentmesh-server -p 8080:8080 \
  -v "/path/to/server.toml:/app/deployments/server.toml:ro" \
  ghcr.io/alanjm0532/agentmesh-server:v0.1.0
```

## Build and Development Commands

```bash
make init         # Download project dependencies and install golangci-lint
make build-cli    # Build the CLI: bin/agentmesh
make build-cli VERSION=v0.1.0 # Embed a CLI release version
make build-server # Generate Wire code and build the server: bin/agentmesh-server
make build        # Build the CLI and HTTP server
make run          # Generate Wire code and start the HTTP server
make run-cli ARGS="--config /path/to/cli.toml ping"
make wire         # Generate the server's wire_gen.go
make generate     # Generate project code (Wire)
make fmt          # Format Go code
make lint         # Generate Wire code and run golangci-lint
make help         # Show available commands
```

`make init` installs golangci-lint v2.15.0 with support for Go 1.27 export data.
Add `$(go env GOPATH)/bin` to your `PATH`
to use it. `.golangci.yml` uses the default golangci-lint v2 rules.
GitHub Actions runs `make init` and `make lint` on branch pushes and pull requests.
Both tag release workflows reuse this check and require it to pass before building.

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

Pushing any Git tag triggers `.github/workflows/release-cli.yml`. After lint passes, the workflow builds the CLI using the Go version declared in `go.mod` and publishes standalone executables to a GitHub Release for that tag after all builds succeed.

| OS | Release Assets |
| --- | --- |
| Linux | `agentmesh_linux_amd64`, `agentmesh_linux_arm64` |
| macOS | `agentmesh_darwin_amd64`, `agentmesh_darwin_arm64` |
| Windows | `agentmesh_windows_amd64.exe`, `agentmesh_windows_arm64.exe` |

Each asset is a single executable. OS and architecture suffixes give every asset a unique name within the release, with `darwin` identifying macOS.
The workflow also publishes `install.sh`, `install.ps1`, `agentmesh-skill.zip`,
and `SHA256SUMS`: ten release assets in total. The ZIP contains `SKILL.md` and
`references/` at its root. The checksum manifest covers the six executables,
installers, and skill archive. Release builds embed their Git tag as the CLI
version, available through `agentmesh --version`.

Both installers download and verify the CLI, then invoke
`agentmesh skill install --version <tag>`. The command verifies and extracts the
skill archive to `~/.agents/skills/agentmesh/`. Download, checksum, or extraction
failures stop installation and preserve the existing skill.

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
