-- +goose Up
CREATE TABLE trip_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    actor_id UUID REFERENCES auth.users(id),
    actor_type VARCHAR(20) NOT NULL,
    payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_trip_events_trip_id ON trip_events(trip_id);
CREATE INDEX idx_trip_events_created_at ON trip_events(created_at);

-- +goose Down
DROP TABLE IF EXISTS trip_events;