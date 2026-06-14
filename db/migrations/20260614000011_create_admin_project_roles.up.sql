CREATE TABLE admin_project_roles (
    admin_user_id UUID NOT NULL REFERENCES admin_users(id),
    project_id UUID NOT NULL REFERENCES projects(id),
    role TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (admin_user_id, project_id)
);
