# Configuration and connectivity

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
