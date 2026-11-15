# Authentication

Use the [CLI reference](cli-reference.md#authentication) for command syntax and
outputs. Reuse any selected `--config` path throughout the workflow.

## Check status and log in

1. Run `agentmesh auth status`. If it succeeds, reuse the existing login. If it
   fails because of a connection error, resolve that through
   [Configuration](configuration.md) and check status again before starting login.
2. When login is needed, use the user's configured server. For a missing or
   explicitly changed server URL, follow Configuration; do not guess the URL.
3. Run `agentmesh auth login`. In a remote or browserless terminal, use
   `--no-browser`.
4. Show the printed authorization code and URL. Ask the user to sign in if needed,
   verify that the code matches, and click **Authorize AgentMesh CLI**. Keep the
   login command running while they authorize; the CLI saves the token.
5. After login completes, run `agentmesh auth status` to verify the account. If
   login expires, start it again and present the new URL and code.

## Log out

Run `agentmesh auth logout` when requested. This removes local credentials and
revokes the current remote token. If the command reports that remote revocation
failed, explain that local credentials were removed and direct the user to revoke
the token from the web token management page.
