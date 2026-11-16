# CLI reference

## Common usage

- `--config "<path>"`: select a configuration file; reuse that path for related commands.
- `--help`: inspect usage for any command, including operations not listed here.
- Success: exit `0`. Report completion only after the command succeeds.
- Failure: exit `1`, error on stderr. Report the actual error rather than assuming success.

## Authentication

| Command | Arguments | Success: exit `0`, stdout |
| --- | --- | --- |
| `agentmesh auth login [--no-browser] [--name <name>]` | Browser approval | Current user and configuration path |
| `agentmesh auth status` | None | Current authenticated user |
| `agentmesh auth logout` | None | Local credential removed; current remote token revoked |

For the browser authorization workflow, see [Authentication](authentication.md).

## Messages

| Command | Arguments | Success: exit `0`, stdout |
| --- | --- | --- |
| `agentmesh message send <user> [text]` | Exactly one of text, `--file <path>`, `--stdin`; optional `--format json`, `--feishu` / `--feishu=false` | Sent message ID, or the JSON message; Feishu status on stderr |
| `agentmesh message recv [user]` | Optional `--follow`, `--duration <minutes>` (default 5, requires `--follow`), `--limit 1..200`, `--format jsonl`, `--output <path>` / `-o <path>` | Unread messages to stdout or appended to the output file; confirms read positions after successful output |
| `agentmesh message history <user>` | Optional `--before <cursor>`, `--limit 1..200`, `--format json` | History page; does not mark messages read |

For interpreting message requests and composing replies, see [Messaging](messaging.md).

### Message data and pagination

- Message bodies: 1–131,072 UTF-8 bytes, no NUL; inline text is never interpreted as a file path.
- History defaults to 50 messages and receive to 100. History JSON returns `next_cursor` and `has_more`; text pagination hints go to stderr.
- Messages are plaintext (`encryption_algorithm: none`). Read state is shared by all CLI installations for the account. Failed acknowledgements may cause repeat output; deduplicate by message ID.

### Receive output and follow mode

- `message recv --output /tmp/messages.jsonl --format jsonl` appends messages to a file, syncing each batch before confirming read positions. New files use `0600` permissions and remain after exit.
- `message recv --follow` polls the server every 5 seconds and fetches backlog pages immediately while more unread messages remain. It exits successfully after 5 minutes by default; use `--duration 30` for 30 minutes. Duration must be a positive integer in minutes and includes requests, retries, and polling pauses. The server returns each query result immediately.

## Feishu notifications

| Command | Arguments | Success: exit `0`, stdout |
| --- | --- | --- |
| `agentmesh feishu login [--no-browser]` | Browser authorization using the configured app credentials | Authorization saved locally |
| `agentmesh feishu bind <user> <email>` | Exact AgentMesh username and Feishu email | Mapping saved in local SQLite |
| `agentmesh feishu bindings` | None | Username and email pairs for the configured server and app |
| `agentmesh feishu unbind <user>` | Exact AgentMesh username | Mapping removed |
| `agentmesh feishu resend <user> <message-id>` | Existing outgoing message in this conversation | Sent to Feishu only |
| `agentmesh feishu logout` | None | Local user tokens removed; authorization grant in Feishu is not revoked |

Read [Configuration](configuration.md#feishu-notifications) for app settings.
The integration is disabled by default. App ID and App Secret come from the CLI
configuration; recipient mappings and user tokens use the CLI's shared SQLite
database through Ent (`database.path`, default `agentmesh.db` beside the config).
Reuse the same `--config` file for all related commands.

If AgentMesh sends successfully but the Feishu notification fails, `message send`
still exits `0` and prints a warning on stderr. Use `feishu resend` to retry only
the notification. Feishu deduplication lasts one hour; a later resend may deliver
another copy. Explicit `feishu` command failures exit `1`.

## Configuration and connectivity

| Command | Arguments | Success: exit `0`, stdout |
| --- | --- | --- |
| `agentmesh config set <key> <value>` | [Settings](configuration.md) | `Saved <key> to <path>` |
| `agentmesh config path` | Optional global `--config <path>`; configuration file need not exist | Absolute path of the selected configuration file |
| `agentmesh ping` | None | `pong` |

For choosing and changing settings, see [Configuration](configuration.md).

## CLI maintenance

| Command | Arguments | Success: exit `0`, stdout |
| --- | --- | --- |
| `agentmesh --version` | None | CLI version |
| `agentmesh skill install` | Optional `--version <tag-or-latest>` | Installed skill version and path |
