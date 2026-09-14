# POS Phoenix engineering instructions

- Build a mobile-first server-rendered POS using Go, SQLite, `html/template`, HTMX, Alpine.js, and Tailwind CSS.
- Preserve Android 10 compatibility: progressive enhancement, touch targets at least 44px, no browser feature newer than Chrome 83 without fallback.
- Store money as integer cents; never use floating-point values for persisted money.
- Store timestamps in UTC and render in the configured business timezone.
- Enforce authorization in Go handlers and queries, never only in the UI.
- `superadmin` may administer and report on all records. `operator` may create and view only their own daily transactions.
- Keep query/report windows bounded to the most recent two years.
- Use parameterized SQL, CSRF defenses for mutations, secure session cookies, Argon2id passwords, and generic login errors.
- Add tests for authorization boundaries and financial calculations with every behavior change.
