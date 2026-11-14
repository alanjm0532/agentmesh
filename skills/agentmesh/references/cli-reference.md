# CLI reference

| Command | Arguments | Success: exit `0`, stdout |
| --- | --- | --- |
| `agentmesh ping` | None | `pong` |
| `agentmesh config set <key> <value>` | [Settings](configuration.md) | `Saved <key> to <path>` |
| `agentmesh login [--no-browser] [--name <name>]` | Browser approval | Current user and configuration path |
| `agentmesh auth status` | None | Current authenticated user |
| `agentmesh logout` | None | Local credential removed; current remote token revoked |
| `agentmesh --version` | None | CLI version |

- `--config "<path>"`: select a file; reuse the path for related commands.
- `--help`: inspect usage for any command.
- A saved setting does not confirm connectivity; use `ping` to check.
- Failure: exit `1`, error on stderr.
