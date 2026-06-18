ALTER TABLE projects ADD COLUMN slug TEXT NOT NULL DEFAULT '';

UPDATE projects SET slug = 'default-project' WHERE slug = '' AND name = 'Default Project';
UPDATE projects SET slug = lower(regexp_replace(regexp_replace(name, '[^a-zA-Z0-9 ]', '', 'g'), '\s+', '-', 'g')) WHERE slug = '';

CREATE UNIQUE INDEX idx_projects_slug ON projects (slug);

ALTER TABLE projects ALTER COLUMN slug SET NOT NULL;
