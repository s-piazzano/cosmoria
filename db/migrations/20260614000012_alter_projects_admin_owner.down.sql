ALTER TABLE projects DROP COLUMN admin_owner_id;
ALTER TABLE projects ADD COLUMN owner_id UUID NOT NULL REFERENCES users(id);
