-- +goose Up
ALTER TABLE trips ADD COLUMN IF NOT EXISTS scheduled_pickup_time TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS trip_waypoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    sequence INTEGER NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    address TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_trip_waypoints_trip ON trip_waypoints(trip_id);

-- +goose Down
DROP TABLE IF EXISTS trip_waypoints;
ALTER TABLE trips DROP COLUMN IF EXISTS scheduled_pickup_time;