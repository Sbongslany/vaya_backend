-- +goose Up
ALTER TABLE payments ADD COLUMN IF NOT EXISTS paystack_reference VARCHAR(255);
ALTER TABLE payments ADD COLUMN IF NOT EXISTS paystack_authorization_url TEXT;

CREATE INDEX IF NOT EXISTS idx_payments_paystack_ref ON payments(paystack_reference);

-- +goose Down
ALTER TABLE payments DROP COLUMN IF EXISTS paystack_authorization_url;
ALTER TABLE payments DROP COLUMN IF EXISTS paystack_reference;