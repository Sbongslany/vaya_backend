-- +goose Up
CREATE TYPE geofence_type AS ENUM ('CITY', 'AIRPORT', 'RESTRICTED', 'SURGE_ZONE');

CREATE TABLE geofences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type geofence_type NOT NULL,
    coordinates TEXT NOT NULL, -- Format: "((lng1 lat1), (lng2 lat2), ...)"
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE driver_zone_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    zone_id UUID NOT NULL REFERENCES geofences(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'WAITING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(driver_id, zone_id)
);

CREATE INDEX idx_geofences_active ON geofences(is_active);
CREATE INDEX idx_driver_zone_driver ON driver_zone_assignments(driver_id);

-- +goose Down
DROP TABLE IF EXISTS driver_zone_assignments;
DROP TABLE IF EXISTS geofences;
DROP TYPE IF EXISTS geofence_type;