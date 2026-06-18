ALTER TABLE files ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE records ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE projects DROP COLUMN multitenancy_enabled;