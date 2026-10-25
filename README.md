# Yupana

Invoice ingestion and reconciliation for Peruvian small businesses: a Go modular monolith on Postgres. Project context lives in [`CONTEXT.md`](CONTEXT.md) and the research in [`docs/`](docs/).

One binary, one image, several processes:

| Command | Process |
|---|---|
| `yupana api` | HTTP API; business modules under `/m/<module>/` |
| `yupana worker` | Background jobs (River, Postgres-backed) |
| `yupana bot` | Telegram bot |
| `yupana migrate` | One-shot: schema migrations as the owner, then the app role password |
| `yupana healthcheck <url>` | Exit 0 when `<url>` answers 200 (used by container health checks) |

## Run it

The stack expects the external `traefik-public` network from `luciel-platform`; for a local run, create it once with `docker network create traefik-public`.

Create `.env` (gitignored) next to `docker-compose.yml`:

| Variable | Required | Meaning |
|---|---|---|
| `YUPANA_OWNER_DB_PASSWORD` | yes | Password of `yupana_owner`, the superuser that owns the schema. Only `migrate` and `backup` use it. |
| `YUPANA_APP_DB_PASSWORD` | yes | Password of `yupana_app`, the role the processes use. It is subject to row-level security and owns nothing. |
| `YUPANA_API_HOST` | yes | Public host name Traefik routes to the API. |
| `YUPANA_TELEGRAM_TOKEN` | no | Bot token; without it the bot only serves its health check. |
| `YUPANA_LOG_LEVEL` | no | `debug`, `info` (default), `warn` or `error`. |
| `YUPANA_IMAGE` | no | Image to run; defaults to the locally built `yupana:local`. |
| `BACKUP_UID`, `BACKUP_GID` | no | Host user that owns `./backups` (default `1000`). |
| `BACKUP_KEEP_DAYS` | no | Days a dump is kept (default `14`). |
| `BACKUP_INTERVAL_SECONDS` | no | Time between dumps (default `86400`). |

Passwords go inside connection URLs, so use URL-safe values such as `openssl rand -hex 32`.

```sh
docker compose up -d --build
docker compose ps
```

`migrate` runs first and exits; `api`, `worker` and `bot` start only after it succeeds.

## Backups and restore

The `backup` service writes `./backups/yupana-<UTC timestamp>.dump` (`pg_dump` custom format) every `BACKUP_INTERVAL_SECONDS` and deletes dumps older than `BACKUP_KEEP_DAYS`. Copy `./backups` off the VPS; a backup on the same disk does not survive losing that disk.

To restore a dump into an empty database:

```sh
docker compose stop api worker bot
docker compose exec -T postgres psql -U yupana_owner -d postgres -c "DROP DATABASE yupana WITH (FORCE)" -c "CREATE DATABASE yupana"
# Grants in the dump name yupana_app; on a fresh cluster create the role first.
docker compose exec -T postgres psql -U yupana_owner -d yupana -c "DO \$\$ BEGIN CREATE ROLE yupana_app LOGIN; EXCEPTION WHEN duplicate_object THEN NULL; END \$\$"
docker compose exec -T postgres pg_restore -U yupana_owner -d yupana --exit-on-error < backups/yupana-<timestamp>.dump
docker compose run --rm migrate   # sets the app role password; already-applied migrations are skipped
docker compose start api worker bot
```

## Develop

Go 1.26. Unit tests run anywhere:

```sh
go test -short ./...
```

Integration tests need a Postgres superuser URL. Each test creates and drops its own database:

```sh
docker run -d --name yupana-testdb -e POSTGRES_PASSWORD=test -p 55432:5432 pgvector/pgvector:0.8.7-pg18-trixie
YUPANA_TEST_DATABASE_URL=postgres://postgres:test@localhost:55432/postgres go test ./...
```
