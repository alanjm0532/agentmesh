# Configuration

```sh
agentmesh config set server "https://example.com"
agentmesh config set timeout "10s"
```

| Key | Value |
| --- | --- |
| `server` | User's complete HTTP/HTTPS URL; ask if unknown. |
| `timeout` | Positive duration with units, e.g. `5s` or `250ms`. |

Change only requested settings. For a selected file, prepend
`--config "<path>"` and reuse it when running `ping`.
See [CLI reference](cli-reference.md) for result checks.

Changing `server` preserves the saved CLI token. Authentication requires HTTPS
except on loopback hosts. See [authentication](../../../docs/authentication.md).
