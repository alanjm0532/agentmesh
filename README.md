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
│   │   ├── ent/schema/         # Ent database schemas
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
│       ├── routes.go           # Shared API route prefix
│       └── ping.go             # PingResponse
├── migrations/                 # Reserved for database migrations
├── docs/
│   ├── config/                 # Server and CLI configuration examples
│   ├── configuration.md        # Configuration setup and database settings
│   └── installation.md         # CLI installation instructions
├── deployments/                # Local configuration, ignored by Git
│   ├── server.toml             # Default server configuration
│   └── cli.toml                # Example CLI configuration
├── go.mod
├── go.sum
├── Dockerfile                  # Multi-stage server image build
├── Makefile
└── README.md
```

`pkg/protocol` defines only request and response types, route constants, and JSON fields used for server–CLI communication, without depending on Gin or CLI libraries.
Both the server and CLI import `github.com/alanjm0532/agentmesh/pkg/protocol`.
Web-only request and response types stay in `internal/server/handler`; Web routes are defined inside `internal/server`.

| Feature | Go Library |
| --- | --- |
| HTTP server | Gin |
| CLI commands | Cobra |
| Configuration | Viper |
| HTTP requests | `github.com/go-resty/resty/v2` |
| JSON | `encoding/json` |
| Database ORM | Ent (MySQL, PostgreSQL, SQLite) |
| Password hashing | `golang.org/x/crypto/bcrypt` |

Use English for documentation, source comments, CLI help, output, errors, and logs.
The web frontend supports English and Simplified Chinese. See [agent.md](agent.md) for project conventions.

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
mkdir -p deployments
cp docs/config/server.toml deployments/server.toml
cp docs/config/cli.toml deployments/cli.toml
make init
make run
```

Update database credentials before starting the server. See
[Configuration examples](docs/configuration.md) for PostgreSQL, MySQL, SQLite,
and CLI configuration files.

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

- `GET /`: serves the AgentMesh homepage with the current account entry.
- `GET /login.html`: logs in, creating new accounts automatically.
- `GET /account.html`: serves token management and logout for the current account.
- `GET /static/*`: serves embedded HTML, CSS, JavaScript, and local vendor assets.
- `GET /api/ping`: returns `{"message":"pong"}`. The CLI's `ping` command calls this endpoint.
- `POST /web/login`: verifies an existing account or creates one if the username does not exist, then establishes a Cookie Session.
- `POST /web/logout`: clears and saves the current session.
- `GET /web/profile`: returns the current user; missing or invalid sessions receive HTTP 401.
- `GET /web/token`: returns the current user's access token; requires a Cookie Session.
- `POST /web/token/reset`: generates and persists a replacement token for the current user; requires a Cookie Session.

Page and static asset routes are registered in the `/` group. Frontend endpoints
use the `/web` group, and CLI endpoints use the `/api` group. Each HTTP client
adds its own route prefix centrally; configure the CLI with the server's root URL.

```bash
curl http://127.0.0.1:8080/api/ping
```

### Web frontend

Open `http://127.0.0.1:8080/` after starting the server. The frontend uses HTML,
local Bootstrap 5.3.3 CSS, and native JavaScript ES Modules with Fetch API.
It requires no frontend framework, Node.js, package manager, or build step.
All frontend files are embedded in the server binary and served by Gin, so the
Docker image needs no separate static asset copy or CDN access at runtime.

```text
internal/server/handler/static/
├── index.html                  # View: homepage and current account entry
├── login.html                  # View: login with automatic account creation
├── account.html                # View: token management and logout
├── css/app.css                 # Application styles
├── images/                     # Brand artwork, favicons, and SVG icon sprite
├── js/
│   ├── main.js                 # Home page entry point
│   ├── access.js               # Account page entry point
│   ├── theme-init.js           # Apply saved or system theme before styles load
│   ├── components/             # UI events and DOM rendering
│   ├── services/               # Business logic
│   ├── api/                    # Endpoint definitions and shared Fetch client
│   ├── locales/                # English and Simplified Chinese translations
│   └── utils/                  # Translation, theme, and notification helpers
└── vendor/bootstrap/           # Local CSS and MIT license
```

Page entry points initialize UI components. Components call services, which use
API modules to access the backend. The shared HTTP client manages JSON requests,
session cookies, network errors, HTTP failures, and invalid JSON responses.
Dynamic messages are rendered with `textContent`.

Both pages include English / Simplified Chinese language buttons. On the first
visit, the UI follows the browser language (Chinese maps to Simplified Chinese;
other languages fall back to English). Manual selections are saved in
`localStorage` and used across pages and future visits. Switching languages updates
page titles, navigation labels, accessibility labels, and
notifications immediately without resetting forms or sending API requests.
Translation keys and parameters are retained for dynamic status and error messages.
API response content remains in the language returned by the server. Translation
resources are local ES Modules; no i18n library or build step is required.

Both pages also offer icon buttons for day, night, and system appearance modes using Bootstrap's
`data-bs-theme` attribute. The default follows the operating system and responds
to system theme changes. Manual selections are saved in `localStorage`, shared
across pages and open tabs, and applied before styles load to avoid a theme flash.
Theme switching preserves form input and works alongside language switching.

The responsive navigation and footer link to
`https://github.com/alanjm0532/agentmesh`. Language and appearance buttons expose
their selected state with `aria-pressed`; appearance buttons include localized
labels and tooltips. Icons are served from a local SVG sprite.

The `/api/ping` and `/web/login`
endpoints are public. Login uses `gin-contrib/sessions` with a
cookie store. The signed `agentmesh_session` cookie contains only the user ID,
uses `HttpOnly` and `SameSite=Lax`, is scoped to `/web`, and expires after one day.
The login page checks `/web/profile` on load and after login to confirm the browser
session. Successful login redirects to the homepage, where the account entry links
to `/account.html`. Already authenticated visitors to the login page go directly
to token management. The account page redirects unauthenticated visitors to login.
It shows the current username and provides a masked token with show/hide, copy,
and reset controls. Reset requires confirmation and replaces the persisted token.
Token responses use `Cache-Control: no-store`. Logout clears and saves the session,
then returns to the homepage.
Web requests send same-origin cookies automatically. Access tokens are stored for
future API use; Bearer token authentication is not implemented.
Protect additional Web routes with `middleware.SessionAuth()`.

Edit files directly under `static/`; changes take effect with the next server
build because the assets are embedded.

The server reads `deployments/server.toml` relative to the working directory by default.
Copy a template from [docs/config](docs/config) and follow the
[configuration setup](docs/configuration.md):

```toml
[server]
addr = "127.0.0.1:8080"
mode = "debug"
session_secret = "replace-with-a-random-secret-at-least-32-characters"
session_secure = false

[database]
driver = "postgres"
dsn = "postgres://agentmesh:password@127.0.0.1:5432/agentmesh?sslmode=disable&timezone=UTC"
```

`server.mode` supports `debug`, `release`, and `test`. Configuration is validated by `Config.Validate()`.
Set `server.session_secret` to a private random value of at least 32 characters;
use the same value across restarts and server instances to retain sessions.
Set `server.session_secure = true` when the browser connects over HTTPS.
After changing the listening address, update `server` in the CLI configuration accordingly.

Set `database.driver` and `database.dsn` to select the database:

| Driver | Example DSN |
| --- | --- |
| `sqlite` | `file:agentmesh.db?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_timezone=UTC` |
| `mysql` | `agentmesh:password@tcp(127.0.0.1:3306)/agentmesh?parseTime=true&charset=utf8mb4&loc=UTC&time_zone=%27%2B00%3A00%27` |
| `postgres` | `postgres://agentmesh:password@127.0.0.1:5432/agentmesh?sslmode=disable&timezone=UTC` |

The MySQL or PostgreSQL database must already exist. The server creates the Ent
schema on startup and closes the database during dependency cleanup. SQLite uses
the pure Go `modernc.org/sqlite` driver, including in the CGO-disabled Docker build.
Its database path is relative to the server's working directory. When using SQLite
in Docker, mount a writable data directory and point the DSN to a file inside it.
For a SQLite in-memory database, use `file::memory:?cache=shared&_pragma=foreign_keys(1)&_timezone=UTC`.

Database timestamps use UTC. Ent generates creation and update times through
`util.NowUTC`, and login responses serialize them with the `Z` UTC suffix.
Database connectors enforce UTC for every new connection, overriding conflicting
DSN time zones: PostgreSQL uses the `timezone=UTC` startup parameter; MySQL uses
`parseTime=true`, driver location `UTC`, and session `time_zone='+00:00'`; SQLite
uses `_timezone=UTC` to normalize timestamp reads and writes. This does not change
the process-wide local time zone. Convert timestamps to the user's local time
when displaying them.

Existing rows are not rewritten. PostgreSQL `timestamptz` values keep their
original instants; their display follows the querying session's time zone.
Historical timestamps that lack an offset or were written with mismatched time
zones need a separate data migration based on their original time zone.

The `users` table contains `id` (UUIDv7 primary key), `username` (unique), `password`
(bcrypt hash), `access_token` (unique), `created_at`, and `updated_at`. Ent sets
timestamps on creation and refreshes `updated_at` for Ent updates. Username
uniqueness is enforced by the database, including concurrent first logins; text
comparison follows the selected database's collation. Passwords are hashed using
`bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)` before insertion.
Each new account uses `util.GenerateToken` to generate an access token with the
`agentmesh_` prefix followed by 64 hexadecimal characters (256 random bits). Login responses
include the user ID, username, and timestamps, and omit the password hash and access token.
Existing accounts require the correct password; concurrent first logins verify the password
of the account that was created first.

```bash
curl -X POST http://127.0.0.1:8080/web/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alan","password":"my-password"}'
```

Successful login returns HTTP 200, including when an account is created automatically.
An incorrect password for an existing username returns HTTP 401.
Malformed JSON and passwords exceeding bcrypt's 72-byte limit return HTTP 400.

Use `-config` to specify a server configuration file:

```bash
make run ARGS="-config /path/to/server.toml"
# Using the compiled server binary
./bin/agentmesh-server -config /path/to/server.toml
```

The server uses Zap at the `info` level by default and writes logs to stdout.
The `release` mode uses JSON encoding; `debug` and `test` use console encoding.
Gin request and recovery logs use `gin-contrib/zap`. Access logs skip `/static`,
`/static/*`, and `/favicon.ico`; page and API requests are still logged.
Panic recovery logging remains enabled for every route.
The cleanup function returned by Wire closes the database and flushes the logger.

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

The Dockerfile generates Ent and Wire code and builds the server with Go 1.27.4.
The runtime image runs as a non-root user and reads the configuration mounted at
`/app/deployments/server.toml` at startup. Configuration files are excluded from
the build context.

Replace `/path/to/server.toml` with the absolute path to your configuration file.
Check the running server with `curl http://127.0.0.1:8080/api/ping`.

### Server image releases

Pushing a Git tag triggers `.github/workflows/release.yml`, which runs lint once.
After lint passes, `.github/workflows/release-server.yml` builds the Linux amd64
server image and pushes it to GitHub Container Registry (GHCR).
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
make build-server # Generate Ent and Wire code and build the server: bin/agentmesh-server
make build        # Build the CLI and HTTP server
make run          # Generate Ent and Wire code and start the HTTP server
make run-cli ARGS="--config /path/to/cli.toml ping"
make ent          # Generate Ent code from internal/server/ent/schema
make wire         # Generate Ent code and Wire injectors for the server and tests
make generate     # Generate project code (Ent and Wire)
make fmt          # Format Go code
make lint         # Generate Ent and Wire code and run golangci-lint
make test         # Generate code and run tests with deployments/server.toml
make help         # Show available commands
```

`make init` installs golangci-lint v2.15.0 with support for Go 1.27 export data.
Add `$(go env GOPATH)/bin` to your `PATH`
to use it. `.golangci.yml` uses the default golangci-lint v2 rules.
GitHub Actions runs `make init` and `make lint` only before tag releases.
Each tag runs this check once, then starts CLI and server image publishing in parallel.

Use the compiled CLI:

```bash
./bin/agentmesh config set server https://agentmesh.example.com
./bin/agentmesh ping
./bin/agentmesh --config /path/to/cli.toml ping
```

Ent and Wire are managed through `tool` directives in `go.mod`; no global installation is required.
After changing dependencies, run `make wire` to generate `internal/server/wire_gen.go`
and `internal/testutil/wire_gen.go`.
Generate Ent and Wire code before running `go run ./cmd/server` directly as well.
Ent output, `wire_gen.go`, and `bin/` build artifacts are ignored by Git. Keep Ent
schema sources under `internal/server/ent/schema/`; do not edit generated files manually.

Tests call `testutil.New(t)` to obtain dependencies assembled by a dedicated Wire
injector in `internal/testutil/wire.go`. The injector reuses the server's repository
and service provider sets with a database client and a silent Zap logger.
It exposes the Ent client, repositories, and services through `Dependencies`:

```go
deps := testutil.New(t)
loggedIn, err := deps.UserService.Login(ctx, username, password)
```

Database connections close automatically after each test. New tests can use these dependencies without
manually constructing repositories and services. Add new dependencies to the
struct and their provider sets, then regenerate Wire code with `make wire`.
The injector calls `config.Load(configPath)` and uses `repository.NewDatabase`
from the shared repository provider set to open and migrate the database.
`testutil.New` registers Wire's cleanup with the test. Configuration loading,
validation, and database initialization are shared with the server.

All tests read `deployments/server.toml` relative to the project root. The shared
`configPath` constant lives in `internal/testutil/dependencies.go`. The helper
walks upward from the test package directory to locate `go.mod`, then passes an
absolute configuration path to Wire without changing the working directory.
Set `database.driver` and `database.dsn` in that file to select PostgreSQL, MySQL,
or SQLite.

```bash
make test
make test TEST_ARGS='-run TestUserServiceLogin -v'
```

Tests use unique account usernames and write directly to the selected database.
Database timestamp precision can vary by driver, so timestamp assertions account
for fractional-second truncation.

## CLI Releases

Pushing any Git tag triggers `.github/workflows/release.yml`. After its shared lint check passes, `.github/workflows/release-cli.yml` builds the CLI using the Go version declared in `go.mod` and publishes standalone executables to a GitHub Release for that tag after all builds succeed.

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
