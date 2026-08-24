-- +goose Up
CREATE TABLE vehicle_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL,
    base_fare NUMERIC(10,2) NOT NULL DEFAULT 0,
    per_km_rate NUMERIC(10,2) NOT NULL DEFAULT 0,
    per_min_rate NUMERIC(10,2) NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE platform_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    commission_percentage NUMERIC(5,2) NOT NULL DEFAULT 20.00,
    cancellation_fee NUMERIC(10,2) NOT NULL DEFAULT 0.00,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed default settings
INSERT INTO platform_settings (commission_percentage, cancellation_fee) VALUES (20.00, 50.00);

-- Seed default vehicle type
INSERT INTO vehicle_types (name, slug, base_fare, per_km_rate, per_min_rate) 
VALUES ('Standard', 'standard', 25.00, 5.00, 1.50);

-- +goose Down
DROP TABLE IF EXISTS platform_settings;
DROP TABLE IF EXISTS vehicle_types;