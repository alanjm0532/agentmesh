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
│       └── config/             # CLI configuration loading with Viper
├── pkg/
│   └── protocol/               # AgentMesh Protocol shared by server and CLI
│       ├── ping.go             # PingResponse
│       └── hello.go            # HelloRequest, HelloResponse
├── migrations/                 # Reserved for database migrations
├── docs/                       # Reserved for documentation
├── deployments/
│   ├── server.toml             # Default server configuration
│   └── cli.toml                # Default CLI configuration
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

Start the HTTP server from the repository root:

```bash
make init
make run
```

The server listens on `127.0.0.1:8080` by default. Run the CLI in another terminal:

```bash
go run ./cmd/agentmesh ping
# Or use the Makefile
make run-cli ARGS="ping"
```

On success, the CLI prints `pong` and exits with code 0. Connection failures, timeouts, unexpected HTTP status codes, and invalid responses produce an error on stderr and exit with code 1.

## CLI Configuration and Flags

```bash
go run ./cmd/agentmesh --help
go run ./cmd/agentmesh ping --help
go run ./cmd/agentmesh --config /path/to/cli.toml ping
```

| Flag | Default | Description |
| --- | --- | --- |
| `--config` | `deployments/cli.toml` | Path to the required CLI TOML configuration file |

Viper reads `deployments/cli.toml` relative to the working directory by default:

```toml
server = "http://127.0.0.1:8080"
timeout = "5s"
```

The configuration file must exist and be readable. Both settings are required:

| Setting | Validation |
| --- | --- |
| `server` | A valid HTTP or HTTPS URL |
| `timeout` | A valid Go duration string greater than zero, such as `5s` |

Runtime settings come exclusively from the TOML file. There are no fallback values, environment variable overrides, or `--server` and `--timeout` flags.
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
./bin/agentmesh ping
./bin/agentmesh --config /path/to/cli.toml ping
```

Wire is managed through the `tool` directive in `go.mod`; no global installation is required.
After changing server dependencies, run `make wire` to generate `internal/server/wire_gen.go`.
Generate Wire code before running `go run ./cmd/server` directly as well.
`wire_gen.go` and `bin/` build artifacts are ignored by Git. Do not edit generated files manually.
