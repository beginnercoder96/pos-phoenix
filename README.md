# POS Phoenix

A portfolio-grade, mobile-first cash-flow POS built with Go, SQLite, server-rendered templates, HTMX, Alpine.js, and Tailwind CSS. It targets Android 10 / Chrome 83 and works with ordinary HTML forms when JavaScript is unavailable.

## Why SQLite

SQLite is the simplest free deployment option: no database server, account, or monthly bill. WAL mode and a single Go process are appropriate for a small portfolio/demo installation. The repository boundary permits a later PostgreSQL adapter when multi-instance scaling is needed.

## Included

- Argon2id password hashing and database-backed opaque sessions
- `superadmin` and `operator` roles
- Daily transaction entry with integer-cent money storage
- Jakarta business-day boundaries with UTC persistence and WIB rendering
- Operator-owned daily views; superadmin global daily, monthly, and custom reports
- Custom report windows bounded to the most recent two years
- CSRF protection and hardened session cookies
- Responsive, touch-friendly dashboard
- Embedded schema initialization, Docker packaging, and starter tests
- Specialized Copilot agents under `.github/agents/`

## Run locally

1. Copy `.env.example` values into your shell and replace the bootstrap password.
2. Run `make setup`.
3. Run `make dev`.
4. Open `http://localhost:8080`.

The bootstrap administrator is created only when no superadmin exists. Do not keep the sample password in a deployed environment. Set `SESSION_SECURE=true` behind HTTPS.

## Agents

Select **POS Orchestrator** for cross-cutting work. It can delegate to:

- **POS Architect** — acceptance criteria and system design
- **POS Backend** — Go handlers and services
- **POS Frontend** — HTMX, Alpine.js, templates, Tailwind
- **POS Data** — SQLite, migrations, reports, financial queries
- **POS Security** — Argon2id, sessions, CSRF, authorization
- **POS Android** — Android 10 / Chrome 83 compatibility
- **POS QA** — tests and release checks

The `/pos-delivery` skill provides a repeatable vertical-slice workflow.

## Next portfolio milestones

1. User/operator administration and password reset
2. Business timezone configuration and bounded custom reports up to two years
3. Pagination and CSV export
4. Audit log and reversible correction workflow
5. PWA manifest/service worker after offline-conflict rules are designed
6. PostgreSQL repository adapter for multi-instance hosting
