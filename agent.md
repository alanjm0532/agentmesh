# Project Conventions

- Use English for source comments, documentation, configuration descriptions, CLI help, output, errors, logs, and server response messages.
- The web frontend supports English and Simplified Chinese through local ES Module translation resources and language buttons. Keep it free of frontend frameworks and build tools.
- Keep `pkg/protocol` limited to protocols shared by the server and CLI. Define Web-only request/response types and route constants inside `internal/server`, outside `pkg/protocol`.
- All database operations, including queries, schema changes, migrations, and transactions, must support MySQL, PostgreSQL, and SQLite. Avoid database-specific SQL or behavior unless equivalent implementations are provided for all three databases.
- Keep HTML, CSS, and JavaScript separate, with small page entry modules and `components -> services -> api` layering; reuse shared HTTP, i18n, and theme helpers.
- Use `confirmAction` with native `<dialog>` for destructive actions and `showNotice` for routine feedback; preserve cancellation, keyboard access, localization, themes, responsive layout, and accessible error fallbacks.
