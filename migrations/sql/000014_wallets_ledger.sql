-- +goose Up
CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE REFERENCES auth.users(id) ON DELETE CASCADE,
    balance NUMERIC(12,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(3) NOT NULL DEFAULT 'ZAR',
    is_platform_wallet BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TYPE ledger_entry_type AS ENUM ('CREDIT', 'DEBIT');
CREATE TYPE ledger_reference_type AS ENUM (
    'TRIP_FARE',
    'PLATFORM_COMMISSION',
    'ADMIN_TOPUP',
    'REFUND',
    'PAYOUT',
    'PAYSTACK_DEPOSIT',
    'PROMOTION_CREDIT'
);

CREATE TABLE ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    entry_type ledger_entry_type NOT NULL,
    amount NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    balance_after NUMERIC(12,2) NOT NULL,
    reference_type ledger_reference_type,
    reference_id UUID,
    description TEXT,
    created_by UUID REFERENCES auth.users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wallets_user ON wallets(user_id);
CREATE INDEX idx_ledger_wallet ON ledger_entries(wallet_id);
CREATE INDEX idx_ledger_reference ON ledger_entries(reference_type, reference_id);
CREATE INDEX idx_ledger_created ON ledger_entries(created_at);

-- Seed platform wallet (no user_id required for platform wallets)
INSERT INTO wallets (id, user_id, balance, currency, is_platform_wallet)
VALUES (
    'a0000000-0000-0000-0000-000000000001',
    NULL,
    0.00,
    'ZAR',
    TRUE
)
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS ledger_entries;
DROP TYPE IF EXISTS ledger_reference_type;
DROP TYPE IF EXISTS ledger_entry_type;
DROP TABLE IF EXISTS wallets;