-- +goose Up
CREATE TABLE IF NOT EXISTS driver_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    document_type VARCHAR(50) NOT NULL,
    file_url TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    rejection_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE auth.users ADD COLUMN IF NOT EXISTS onboarding_status VARCHAR(20) NOT NULL DEFAULT 'NOT_STARTED';

CREATE INDEX IF NOT EXISTS idx_driver_documents_user ON driver_documents(user_id);
CREATE INDEX IF NOT EXISTS idx_driver_documents_status ON driver_documents(status);

-- +goose Down
DROP TABLE IF EXISTS driver_documents;
ALTER TABLE auth.users DROP COLUMN IF EXISTS onboarding_status;