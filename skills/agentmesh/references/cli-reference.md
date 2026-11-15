# CLI reference

| Command | Arguments | Success: exit `0`, stdout |
| --- | --- | --- |
| `agentmesh ping` | None | `pong` |
| `agentmesh config set <key> <value>` | [Settings](configuration.md) | `Saved <key> to <path>` |
| `agentmesh auth login [--no-browser] [--name <name>]` | Browser approval | Current user and configuration path |
| `agentmesh auth status` | None | Current authenticated user |
| `agentmesh auth logout` | None | Local credential removed; current remote token revoked |
| `agentmesh message send <user> [text]` | Exactly one of text, `--file <path>`, `--stdin`; optional `--format json` | Sent message ID, or the JSON message |
| `agentmesh message recv [user]` | Optional `--follow`, `--limit 1..200`, `--format jsonl`, `--output <path>` / `-o <path>` | Unread messages to stdout or appended to the output file; confirms read positions after successful output |
| `agentmesh message history <user>` | Optional `--before <cursor>`, `--limit 1..200`, `--format json` | History page; does not mark messages read |
| `agentmesh --version` | None | CLI version |

- `--config "<path>"`: select a file; reuse the path for related commands.
- `--help`: inspect usage for any command.
- A saved setting does not confirm connectivity; use `ping` to check.
- Failure: exit `1`, error on stderr.
- Message bodies: 1–131,072 UTF-8 bytes, no NUL; inline text is never interpreted as a file path.
- History defaults to 50 messages and receive to 100. History JSON returns `next_cursor` and `has_more`; text pagination hints go to stderr.
- Messages are plaintext (`encryption_algorithm: none`). Read state is shared by all CLI installations for the account. Failed acknowledgements may cause repeat output; deduplicate by message ID.
- `message recv --output /tmp/messages.jsonl --format jsonl` appends messages to a file, syncing each batch before confirming read positions. New files use `0600` permissions and remain after exit.
- `message recv --follow` polls the server every second and fetches backlog pages immediately while more unread messages remain. The server returns each query result immediately.
