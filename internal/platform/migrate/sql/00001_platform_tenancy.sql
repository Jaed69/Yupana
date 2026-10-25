-- Tenancy foundation: shared schema + tenant_id + row-level security.
-- Migrations run as the schema owner; the application connects as yupana_app,
-- which owns nothing and has neither SUPERUSER nor BYPASSRLS.

-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'yupana_app') THEN
        CREATE ROLE yupana_app LOGIN NOSUPERUSER NOBYPASSRLS NOCREATEDB NOCREATEROLE;
    END IF;
END
$$;
-- +goose StatementEnd

CREATE SCHEMA platform;
GRANT USAGE ON SCHEMA platform TO yupana_app;

-- Fail closed: an unset or empty app.tenant_id yields NULL, and NULL matches no row.
CREATE FUNCTION platform.current_tenant_id() RETURNS uuid
    LANGUAGE sql STABLE PARALLEL SAFE
    AS $$ SELECT nullif(current_setting('app.tenant_id', true), '')::uuid $$;

-- Every tenant-scoped table is protected through this helper, so the policy
-- shape (ENABLE + FORCE, USING + WITH CHECK) stays identical across modules.
-- +goose StatementBegin
CREATE FUNCTION platform.enable_tenant_rls(tbl regclass, tenant_column name DEFAULT 'tenant_id') RETURNS void
    LANGUAGE plpgsql AS $$
BEGIN
    EXECUTE format('ALTER TABLE %s ENABLE ROW LEVEL SECURITY', tbl);
    EXECUTE format('ALTER TABLE %s FORCE ROW LEVEL SECURITY', tbl);
    EXECUTE format(
        'CREATE POLICY tenant_isolation ON %s USING (%I = platform.current_tenant_id()) WITH CHECK (%I = platform.current_tenant_id())',
        tbl, tenant_column, tenant_column);
END
$$;
-- +goose StatementEnd

CREATE TABLE platform.tenants (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name       text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
SELECT platform.enable_tenant_rls('platform.tenants', 'id');
GRANT SELECT ON platform.tenants TO yupana_app;

-- Which modules each tenant has switched on. Provisioning writes it as the owner.
CREATE TABLE platform.tenant_modules (
    tenant_id  uuid NOT NULL REFERENCES platform.tenants (id) ON DELETE CASCADE,
    module     text NOT NULL,
    enabled_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, module)
);
SELECT platform.enable_tenant_rls('platform.tenant_modules');
GRANT SELECT ON platform.tenant_modules TO yupana_app;

-- +goose Down
DROP SCHEMA platform CASCADE;
