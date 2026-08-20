package postgres

import (
	"context"
	"errors"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

type rowScanner interface {
	Scan(dest ...any) error
}

const tripColumns = `id, passenger_id, driver_id, vehicle_id, trip_type, status, start_pin,
	pickup_latitude, pickup_longitude, pickup_address,
	dropoff_latitude, dropoff_longitude, dropoff_address,
	estimated_fare, final_fare, currency, distance_km,
	long_distance_type, scheduled_departure, scheduled_return, trip_duration_days,
	created_at, updated_at`

func scanTrip(rs rowScanner) (*entities.Trip, error) {
	t := &entities.Trip{}
	if err := rs.Scan(
		&t.ID, &t.PassengerID, &t.DriverID, &t.VehicleID, &t.TripType, &t.Status, &t.StartPIN,
		&t.PickupLatitude, &t.PickupLongitude, &t.PickupAddress,
		&t.DropoffLatitude, &t.DropoffLongitude, &t.DropoffAddress,
		&t.EstimatedFare, &t.FinalFare, &t.Currency, &t.DistanceKM,
		&t.LongDistanceType, &t.ScheduledDeparture, &t.ScheduledReturn, &t.TripDurationDays,
		&t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return t, nil
}

type TripRepository struct {
	pool *pgxpool.Pool
}

func NewTripRepository(pool *pgxpool.Pool) *TripRepository {
	return &TripRepository{pool: pool}
}

func (r *TripRepository) Create(ctx context.Context, trip *entities.Trip) error {
	query := `INSERT INTO trips (` + tripColumns + `)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)`

	_, err := r.pool.Exec(ctx, query,
		trip.ID, trip.PassengerID, trip.DriverID, trip.VehicleID, trip.TripType, trip.Status, trip.StartPIN,
		trip.PickupLatitude, trip.PickupLongitude, trip.PickupAddress,
		trip.DropoffLatitude, trip.DropoffLongitude, trip.DropoffAddress,
		trip.EstimatedFare, trip.FinalFare, trip.Currency, trip.DistanceKM,
		trip.LongDistanceType, trip.ScheduledDeparture, trip.ScheduledReturn, trip.TripDurationDays,
		trip.CreatedAt, trip.UpdatedAt,
	)
	return err
}

func (r *TripRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Trip, error) {
	query := `SELECT ` + tripColumns + ` FROM trips WHERE id = $1`

	trip, err := scanTrip(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return trip, nil
}

func (r *TripRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entities.TripStatus) error {
	query := `UPDATE trips SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, id)
	return err
}

func (r *TripRepository) UpdateStatusAndFinalFare(ctx context.Context, id uuid.UUID, status entities.TripStatus, finalFare float64) error {
	query := `UPDATE trips SET status = $1, final_fare = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, status, finalFare, id)
	return err
}

func (r *TripRepository) FindNearbyRequested(ctx context.Context, lat, lng, radiusKM float64, limit int) ([]*entities.Trip, error) {
	latDelta := radiusKM / 111.0
	lngDelta := radiusKM / (111.0 * math.Cos(lat*math.Pi/180.0))

	query := `SELECT ` + tripColumns + `
		FROM trips
		WHERE status = $1
			AND pickup_latitude BETWEEN $2 AND $3
			AND pickup_longitude BETWEEN $4 AND $5
		ORDER BY created_at ASC
		LIMIT $6`

	rows, err := r.pool.Query(ctx, query,
		entities.StatusRequested,
		lat-latDelta, lat+latDelta,
		lng-lngDelta, lng+lngDelta,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trips []*entities.Trip
	for rows.Next() {
		trip, err := scanTrip(rows)
		if err != nil {
			return nil, err
		}
		if haversineDistanceKM(lat, lng, trip.PickupLatitude, trip.PickupLongitude) <= radiusKM {
			trips = append(trips, trip)
		}
	}
	return trips, rows.Err()
}

func (r *TripRepository) FindActiveByPassengerID(ctx context.Context, passengerID uuid.UUID) (*entities.Trip, error) {
	query := `SELECT ` + tripColumns + `
		FROM trips
		WHERE passenger_id = $1
			AND status = ANY($2)
		ORDER BY created_at DESC
		LIMIT 1`

	activeStatuses := []string{
		string(entities.StatusRequested),
		string(entities.StatusSearchingDrivers),
		string(entities.StatusOffersReceived),
		string(entities.StatusDriverSelected),
		string(entities.StatusDriverAssigned),
		string(entities.StatusDriverConfirmed),
		string(entities.StatusScheduled),
		string(entities.StatusDriverEnRoute),
		string(entities.StatusDriverArrived),
		string(entities.StatusTripStartPending),
		string(entities.StatusTripStarted),
		string(entities.StatusTripInProgress),
		string(entities.StatusOutboundInProgress),
		string(entities.StatusDestinationReached),
		string(entities.StatusDriverRetained),
		string(entities.StatusReturnScheduled),
		string(entities.StatusReturnStarted),
		string(entities.StatusReturnInProgress),
		string(entities.StatusFinalDestination),
		string(entities.StatusQuoteGenerated),
		string(entities.StatusArrivedAtDest),
	}

	trip, err := scanTrip(r.pool.QueryRow(ctx, query, passengerID, activeStatuses))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return trip, nil
}

func (r *TripRepository) AssignDriver(ctx context.Context, tripID, driverID uuid.UUID, status entities.TripStatus) error {
	query := `UPDATE trips SET driver_id = $1, status = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, driverID, status, tripID)
	return err
}

func (r *TripRepository) FindOpenLongDistanceTrips(ctx context.Context, limit int) ([]*entities.Trip, error) {
	query := `SELECT ` + tripColumns + `
		FROM trips
		WHERE trip_type = $1
			AND status = ANY($2)
		ORDER BY scheduled_departure ASC
		LIMIT $3`

	openStatuses := []string{
		string(entities.StatusQuoteGenerated),
		string(entities.StatusSearchingDrivers),
	}

	rows, err := r.pool.Query(ctx, query, entities.TripTypeLongDistance, openStatuses, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trips []*entities.Trip
	for rows.Next() {
		trip, err := scanTrip(rows)
		if err != nil {
			return nil, err
		}
		trips = append(trips, trip)
	}
	return trips, rows.Err()
}

func haversineDistanceKM(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLng := (lng2 - lng1) * math.Pi / 180.0
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}