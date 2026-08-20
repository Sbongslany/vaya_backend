-- +goose Up
CREATE TYPE trip_type AS ENUM ('NORMAL', 'LONG_DISTANCE');
CREATE TYPE trip_status AS ENUM (
    'REQUESTED', 'SEARCHING_DRIVERS', 'OFFERS_RECEIVED', 'DRIVER_SELECTED', 
    'DRIVER_ASSIGNED', 'DRIVER_EN_ROUTE', 'DRIVER_ARRIVED', 'TRIP_START_PENDING', 
    'TRIP_STARTED', 'TRIP_IN_PROGRESS', 'ARRIVED_AT_DESTINATION', 'TRIP_COMPLETED', 
    'PAYMENT_PROCESSING', 'PAYMENT_COMPLETED', 'RATING_PENDING', 'CLOSED',
    'CANCELLED_BY_PASSENGER', 'CANCELLED_BY_DRIVER', 'CANCELLED_BY_SYSTEM'
);
CREATE TYPE offer_type AS ENUM ('NORMAL_FARE', 'OFFER');
CREATE TYPE offer_status AS ENUM ('PENDING', 'ACCEPTED', 'REJECTED', 'EXPIRED');

CREATE TABLE trips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    passenger_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    driver_id UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    vehicle_id UUID REFERENCES auth.vehicles(id) ON DELETE SET NULL,
    
    trip_type trip_type NOT NULL DEFAULT 'NORMAL',
    status trip_status NOT NULL DEFAULT 'REQUESTED',
    
    pickup_latitude DOUBLE PRECISION NOT NULL,
    pickup_longitude DOUBLE PRECISION NOT NULL,
    pickup_address TEXT NOT NULL,
    dropoff_latitude DOUBLE PRECISION NOT NULL,
    dropoff_longitude DOUBLE PRECISION NOT NULL,
    dropoff_address TEXT NOT NULL,
    
    estimated_fare NUMERIC(10,2) NOT NULL,
    final_fare NUMERIC(10,2),
    currency VARCHAR(3) NOT NULL DEFAULT 'ZAR',
    distance_km NUMERIC(10,2),
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE trip_offers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    driver_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    offer_type offer_type NOT NULL DEFAULT 'NORMAL_FARE',
    offered_fare NUMERIC(10,2) NOT NULL,
    status offer_status NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(trip_id, driver_id)
);

CREATE INDEX idx_trips_passenger ON trips(passenger_id);
CREATE INDEX idx_trips_driver ON trips(driver_id);
CREATE INDEX idx_trips_status ON trips(status);
CREATE INDEX idx_trip_offers_trip ON trip_offers(trip_id);
CREATE INDEX idx_trip_offers_driver ON trip_offers(driver_id);

-- +goose Down
DROP TABLE IF EXISTS trip_offers;
DROP TABLE IF EXISTS trips;
DROP TYPE IF EXISTS offer_status;
DROP TYPE IF EXISTS offer_type;
DROP TYPE IF EXISTS trip_status;
DROP TYPE IF EXISTS trip_type;