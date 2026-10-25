# Phase 0 — Go skeleton

Objective: runnable modular-monolith skeleton (api/worker/bot from one image) with tenant isolation, per-tenant modules, a Postgres-backed outbox/queue, hardened Compose, backups and arm64 CI.
Branch: `feat/phase-0-skeleton`. Delivery strategy: `ask-on-risk` (chain strategy asked before the first PR; forecast ~1,500 authored lines). Test runner: `go test ./...` (integration tests need `YUPANA_TEST_DATABASE_URL`; `-short` skips them).

## Specs

Source: `docs/research/informe.md` phases table, row 0: "Monolito FastAPI + Postgres con `tenant_id` y RLS, módulos y `tenant_modules`, outbox con worker, Docker Compose (api/worker/bot) en el VPS, backups" — "Sin aislamiento y sin cola, todo lo demás es deuda". FastAPI is superseded by Go (`CONTEXT.md` decisions table).

- **S1 — Processes.** One Go module, one image, three entrypoints: `api`, `worker`, `bot`. Each starts from env config, logs with `slog` (JSON), exposes a health check and shuts down gracefully on SIGTERM.
- **S2 — Tenant isolation.** Shared schema + `tenant_id` + RLS: `ENABLE` + `FORCE ROW LEVEL SECURITY`, policies with `USING` and `WITH CHECK`, tenant set per transaction with `SET LOCAL` (`set_config(..., true)`), app connects as a non-owner role (migrations run as the owner). Fail closed: with no tenant set, queries return zero rows. Tests prove: zero rows without tenant; tenant A cannot read tenant B; inserting a row for another tenant is rejected.
- **S3 — Modules and per-tenant activation.** Modules live under `internal/modules/<module>`: the root package is the public interface, internals go in `internal/modules/<module>/internal/` (compiler-enforced, so no module imports another's internals). `platform.tenant_modules` holds activation; a request to a module not active for the tenant returns **403**.
- **S4 — Outbox / queue.** Postgres queue with `SKIP LOCKED`, `LISTEN/NOTIFY` only as wake-up (River). Jobs are enqueued in the same transaction as the business change: a rolled-back transaction leaves no job; a committed one is processed by the `worker`.
- **S5 — Deploy.** Docker Compose for the existing VPS: postgres (pgvector, pg18), one-shot `migrate`, `api`, `worker`, `bot`; hardened like `luciel-platform` (`no-new-privileges`, `cap_drop: ALL`, read-only root fs, tmpfs, `mem_limit`), `api` behind Traefik via the external `traefik-public` network.
- **S6 — Backups.** Scheduled `pg_dump` (custom format) with retention, plus a documented restore command.
- **S7 — CI.** GitHub Actions on native `ubuntu-24.04-arm`: `go vet`, tests against a Postgres service, image build. Assumption: GHCR push/Trivy/deploy are left for when the repo has a remote (no remote exists yet).

## Tasks

| ID | Specs | Route | Status | Commit |
|---|---|---|---|---|
| T1 | S1 | inline | done (RED: build failed; GREEN: go test ./... ok, binary smoke healthcheck exit 0) | 1419d51 |
| T2 | S2 | inline (Risk item 2: own commit + verifier) | done (RED: build failed; GREEN: 7 RLS integration tests ok; mutation `WITH CHECK (true)` and dropped FORCE both caught); verifier PASS, no blockers (L6) | d277695 |
| T3 | S3 | inline (Risk item 2: access guard) | done (RED: build failed; GREEN: 6 route cases 200/403/403/401/401/404 + prod ignores dev header); verifier pending | a0f1c55 |
| T4 | S4 | inline | done (RED: build failed; GREEN: rollback leaves 0 jobs, commit leaves 1 and a worker on the app role works it inside the job tenant scope; app role cannot rewrite river_migration; smoke: worker boots as yupana_app with 0 errors) | see L7 |
| T5 | S5, S6 | inline | pending | |
| T6 | S7 | inline (Risk item 5) | pending | |

Route evidence: no Writer trigger (units depend on each other sequentially; no parallel disjoint units), so the parent writes inline.

## Log

- **L1 (2026-10-25, user):** "S i dale" — answer to "¿Arrancamos con la Fase 0?".
- L2: research docs committed first as `c8518fb` on the feature branch.
- L3: T1 committed `1419d51`; assess medium, `under_budget` (slice closed before the high-risk T2).
- L4: T2. Finding: a `FOR ALL` policy without `WITH CHECK` reuses `USING` for writes, so removing it is not a weakening; the meaningful mutation is `WITH CHECK (true)`, which the write tests catch. Owner credentials stay out of `config.Config`; only `yupana migrate` reads `YUPANA_MIGRATE_DATABASE_URL` and sets the `yupana_app` password from `YUPANA_APP_DB_PASSWORD`.
- L5: T3. S3 path refined to `internal/modules/<module>` with Go `internal/` for module internals (agent-authored spec, not a user string). Tenant resolution: no auth exists yet, so prod uses `DenyResolver` (every module route 401) and only `YUPANA_ENV=dev` trusts the `X-Yupana-Tenant` header. Real resolution (bot chat -> tenant, API auth) is a later phase.
- L6: T2 independent verifier: S2 PASS on every clause, no blockers. Probes held: no-tenant writes, `SET ROLE`, `row_security=off`, malformed tenant, policy tampering, object creation, quoted password injection, repeated migrate. Advisories carried forward: (1) with FORCE RLS a non-superuser owner cannot provision tenants, so provisioning needs a superuser/BYPASSRLS role or a SECURITY DEFINER function (decide in T5); (2) in PG16+ a CREATEROLE owner that did not create `yupana_app` cannot set its password; (3) revoke PUBLIC EXECUTE on `platform.enable_tenant_rls`; (4) app role keeps TEMP; (5) only `WithTenant` may set the tenant (session-level `set_config` would leak across the pool); (6) Down does not drop the cluster-wide role (intentional).
- L7: T4. River tables live in schema `river`, migrated by `yupana migrate` as the owner; the app role gets job DML but only SELECT on `river_migration`. The River reindexer is disabled because REINDEX needs table ownership. Phase 0 registers one `platform.ping` worker because River refuses to start without workers; modules add real jobs later.
