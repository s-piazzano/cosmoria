ALTER TABLE projects DROP COLUMN owner_id;
ALTER TABLE projects ADD COLUMN admin_owner_id UUID NOT NULL REFERENCES admin_users(id);
