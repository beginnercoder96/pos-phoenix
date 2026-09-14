---
name: pos-delivery
description: "Use when: planning, implementing, reviewing, or releasing a POS Phoenix feature across architecture, Go, HTMX, SQLite, auth, Android, and tests."
---

# POS delivery workflow

1. Restate the feature as user-visible acceptance criteria.
2. Identify role access for `superadmin` and `operator`; deny by default.
3. Define schema/query implications, including integer cents, UTC timestamps, ownership, indexes, and two-year retention/report boundaries.
4. Implement a thin vertical slice: migration/query, Go service/handler, HTML partial/page, then progressive enhancement.
5. Validate on narrow Android-sized viewport and without JavaScript.
6. Add unit/integration tests for success, invalid input, authentication, authorization, and money totals.
7. Run formatting, CSS build, tests, and Go build.
8. Return changed files, trade-offs, risks, and next action.

Do not silently broaden scope. Prefer standard library and small, maintained dependencies. Never authorize solely by hiding UI controls.
