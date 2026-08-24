-- +goose Up
CREATE TABLE admin_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID,
    details TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_admin_audit_logs_admin ON admin_audit_logs(admin_id);
CREATE INDEX idx_admin_audit_logs_resource ON admin_audit_logs(resource_type, resource_id);

-- +goose Down
DROP TABLE IF EXISTS admin_audit_logs;