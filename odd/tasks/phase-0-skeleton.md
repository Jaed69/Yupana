# Phase 0 — Go skeleton

Objective: runnable modular-monolith skeleton (api/worker/bot from one image) with tenant isolation, per-tenant modules, a Postgres-backed outbox/queue, hardened Compose, backups and arm64 CI.
Branch: `feat/phase-0-skeleton`. Delivery strategy: `ask-on-risk` (chain strategy asked before the first PR; forecast ~1,500 authored lines). Test runner: `go test ./...` (integration tests need `YUPANA_TEST_DATABASE_URL`; `-short` skips them).

## Specs

Source: `docs/research/informe.md` phases table, row 0: "Monolito FastAPI + Postgres con `tenant_id` y RLS, módulos y `tenant_modules`, outbox con worker, Docker Compose (api/worker/bot) en el VPS, backups" — "Sin aislamiento y sin cola, todo lo demás es deuda". FastAPI is superseded by Go (`CONTEXT.md` decisions table).

- **S1 — Processes.** One Go module, one image, three entrypoints: `api`, `worker`, `bot`. Each starts from env config, logs with `slog` (JSON), exposes a health check and shuts down gracefully on SIGTERM.
- **S2 — Tenant isolation.** Shared schema + `tenant_id` + RLS: `ENABLE` + `FORCE ROW LEVEL SECURITY`, policies with `USING` and `WITH CHECK`, tenant set per transaction with `SET LOCAL` (`set_config(..., true)`), app connects as a non-owner role (migrations run as the owner). Fail closed: with no tenant set, queries return zero rows. Tests prove: zero rows without tenant; tenant A cannot read tenant B; inserting a row for another tenant is rejected.
- **S3 — Modules and per-tenant activation.** Modules live under `internal/<module>`, expose a public interface, never import another module's internals. `platform.tenant_modules` holds activation; a request to a module not active for the tenant returns **403**.
- **S4 — Outbox / queue.** Postgres queue with `SKIP LOCKED`, `LISTEN/NOTIFY` only as wake-up (River). Jobs are enqueued in the same transaction as the business change: a rolled-back transaction leaves no job; a committed one is processed by the `worker`.
- **S5 — Deploy.** Docker Compose for the existing VPS: postgres (pgvector, pg18), one-shot `migrate`, `api`, `worker`, `bot`; hardened like `luciel-platform` (`no-new-privileges`, `cap_drop: ALL`, read-only root fs, tmpfs, `mem_limit`), `api` behind Traefik via the external `traefik-public` network.
- **S6 — Backups.** Scheduled `pg_dump` (custom format) with retention, plus a documented restore command.
- **S7 — CI.** GitHub Actions on native `ubuntu-24.04-arm`: `go vet`, tests against a Postgres service, image build. Assumption: GHCR push/Trivy/deploy are left for when the repo has a remote (no remote exists yet).

## Tasks

| ID | Specs | Route | Status | Commit |
|---|---|---|---|---|
| T1 | S1 | inline | done (RED: build failed; GREEN: go test ./... ok, binary smoke healthcheck exit 0) | see L3 |
| T2 | S2 | inline (Risk item 2: own commit + verifier) | pending | |
| T3 | S3 | inline | pending | |
| T4 | S4 | inline | pending | |
| T5 | S5, S6 | inline | pending | |
| T6 | S7 | inline (Risk item 5) | pending | |

Route evidence: no Writer trigger (units depend on each other sequentially; no parallel disjoint units), so the parent writes inline.

## Log

- **L1 (2026-10-25, user):** "S i dale" — answer to "¿Arrancamos con la Fase 0?".
- L2: research docs committed first as `c8518fb` on the feature branch.
