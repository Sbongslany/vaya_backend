-- +goose Up
CREATE TYPE payout_status AS ENUM ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED', 'CANCELLED');

CREATE TABLE payouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    amount NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'ZAR',
    status payout_status NOT NULL DEFAULT 'PENDING',
    bank_name VARCHAR(255) NOT NULL,
    bank_account_number VARCHAR(20) NOT NULL,
    bank_account_name VARCHAR(255) NOT NULL,
    paystack_transfer_reference VARCHAR(255),
    paystack_transfer_id VARCHAR(255),
    failure_reason TEXT,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payouts_user ON payouts(user_id);
CREATE INDEX idx_payouts_status ON payouts(status);

-- Add PAYOUT to ledger reference types if not already there
-- (Already exists from Phase 12 migration)

-- +goose Down
DROP TABLE IF EXISTS payouts;
DROP TYPE IF EXISTS payout_status;