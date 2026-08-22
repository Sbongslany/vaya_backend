-- +goose Up
ALTER TABLE trips ADD COLUMN IF NOT EXISTS route_polyline TEXT;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS route_duration_minutes INTEGER;
ALTER TABLE trips ADD COLUMN IF NOT EXISTS route_distance_km NUMERIC(10,2);

-- +goose Down
ALTER TABLE trips DROP COLUMN IF EXISTS route_distance_km;
ALTER TABLE trips DROP COLUMN IF EXISTS route_duration_minutes;
ALTER TABLE trips DROP COLUMN IF EXISTS route_polyline;