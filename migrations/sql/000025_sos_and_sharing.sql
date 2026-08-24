-- +goose Up
CREATE TYPE sos_status AS ENUM ('ACTIVE', 'RESOLVED', 'FALSE_ALARM');

CREATE TABLE sos_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    triggered_by UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    status sos_status NOT NULL DEFAULT 'ACTIVE',
    triggered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    resolved_by UUID REFERENCES auth.users(id) ON DELETE SET NULL
);

CREATE TABLE trip_share_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    token VARCHAR(64) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sos_alerts_trip ON sos_alerts(trip_id);
CREATE INDEX idx_sos_alerts_status ON sos_alerts(status);
CREATE INDEX idx_trip_share_tokens_token ON trip_share_tokens(token);

-- +goose Down
DROP TABLE IF EXISTS trip_share_tokens;
DROP TABLE IF EXISTS sos_alerts;
DROP TYPE IF EXISTS sos_status;