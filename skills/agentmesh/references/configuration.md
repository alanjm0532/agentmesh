# Configuration and connectivity

## Configuration file location

Run `agentmesh config path` to print the absolute path of the configuration file
selected by the CLI. `agentmesh --config ./work.toml config path` prints the
absolute path of the specified file. This command works even if the file does
not exist or cannot be parsed.

## Settings

```sh
agentmesh config set server "https://example.com"
agentmesh config set timeout "10s"
```

| Key | Value |
| --- | --- |
| `server` | User's complete HTTP/HTTPS URL; ask if unknown. |
| `timeout` | Positive duration with units, e.g. `5s` or `250ms`. |

Change only requested settings. Changing `server` preserves the saved CLI token.
For a selected file, prepend `--config "<path>"` and reuse it for related commands.

## Check connectivity

Run `agentmesh ping` when asked to check connectivity or verify a newly configured
server. A saved setting does not confirm connectivity; a successful ping prints
`pong`.

Use the [CLI reference](cli-reference.md#configuration-and-connectivity) for command
syntax and outputs. For login or an invalid token, see
[Authentication](authentication.md).

## Feishu notifications

Edit the selected CLI TOML configuration to add:

```toml
[database]
path = "agentmesh.db"

[feishu]
enabled = true
app_id = "cli_your_app_id"
app_secret = "your_app_secret"
redirect_uri = "http://localhost:18888/callback"
```

Read App ID and App Secret from this file; do not embed shared credentials in an
executable. Keep the configuration private. In the Feishu developer console,
enable the app's bot capability and the `im:message`,
`im:message.send_as_user`, and `offline_access` permissions, publish the app, and
allow the configured redirect URL. App availability and communication rules
must permit the intended recipients.

Use `agentmesh feishu login` to authorize the sending user. The CLI uses PKCE and
a local browser callback. Add recipient mappings with
`agentmesh feishu bind <user> <email>`; mappings are kept in a new
`feishu_account_mappings` SQLite table, not in the configuration file.
SQLite is shared CLI storage managed through Ent. `database.path` defaults to
`agentmesh.db` beside the selected configuration file; relative paths use that
directory. The database also stores user tokens and is created with `0600`
permissions; values are not encrypted. Authorizations are isolated by the
absolute configuration filename and App ID, while mappings are scoped by server
and App ID. Moving the configuration file requires authorizing again.

The default is disabled. `message send --feishu` or `--feishu=false` overrides the
configured setting for one send. Reuse the same `--config` selection across
authorization, binding, sending, and resending. See the
[CLI reference](cli-reference.md#feishu-notifications) for commands and partial
delivery behavior.
