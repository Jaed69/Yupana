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
| T3 | S3 | inline (Risk item 2: access guard) | done (RED: build failed; GREEN: 6 route cases 200/403/403/401/401/404 + prod ignores dev header); verifier PASS (L8) | a0f1c55 |
| T4 | S4 | inline | done (RED: build failed; GREEN: rollback leaves 0 jobs, commit leaves 1 and a worker on the app role works it inside the job tenant scope; app role cannot rewrite river_migration; smoke: worker boots as yupana_app with 0 errors); verifier PASS, review approved (L8) | cce3c44 |
| T5 | S5, S6 | inline (Risk item 5) | done (no runnable RED: deploy config; checks: `compose config` ok, image 24.6 MB, full stack healthy, backups written, retention prunes only old `yupana-*.dump`, documented restore end to end keeps data and FORCE RLS) | see L9 |
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
- L8: T3+T4 independent verifier: S3 and S4 PASS, no blockers (path tricks never reach the handler; app role denied every DDL/TRUNCATE/REINDEX on `river`; worker soak 0 errors). RDD review of slice 1419d51..cce3c44 (lenses risk, reliability): consent granted, approved, acknowledged (lineage review-bcbbb6b7b8534af9). Advisory follow-ups: queue permission test now asserts 42501 (fixed); `river_job` has no RLS, so job args stay minimal (tenant ID plus references); `ALTER ROLE ... PASSWORD` text may reach server logs if `log_statement` covers DDL; api boot requires the DB; worker drain on stop is untested; dev `HeaderResolver` accepts non-canonical UUID spellings. Carry to T5: tenant provisioning role (FORCE RLS blocks a non-superuser owner) and Postgres logging settings.
- L9: T5. The Postgres owner `yupana_owner` is the image superuser, used only by `migrate` and `backup`. That settles the provisioning advisory for Phase 0: FORCE RLS does not bind a superuser. Postgres keeps the image default `log_statement=none`, so the `ALTER ROLE ... PASSWORD` text is not logged. Measured idle memory on the local stack: api 4 MiB, worker 6 MiB, bot 3.5 MiB, postgres 76 MiB. `.env.example` was not written because the permission settings deny `.env*` paths; the variables are documented in the README table instead. The local `traefik-public` network was left in place for later local runs.
