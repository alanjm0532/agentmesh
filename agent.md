# Project Conventions

- Use English for source comments, documentation, configuration descriptions, CLI help, output, errors, logs, and server response messages.
- The web frontend supports English and Simplified Chinese through local ES Module translation resources and language buttons. Keep it free of frontend frameworks and build tools.
- Keep `pkg/protocol` limited to protocols shared by the server and CLI. Define Web-only request/response types and route constants inside `internal/server`, outside `pkg/protocol`.
