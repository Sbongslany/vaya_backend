package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

const tripOfferColumns = `id, trip_id, driver_id, offer_type, offered_fare, status, created_at, updated_at`

func scanTripOffer(rs rowScanner) (*entities.TripOffer, error) {
	o := &entities.TripOffer{}
	if err := rs.Scan(
		&o.ID, &o.TripID, &o.DriverID, &o.OfferType, &o.OfferedFare, &o.Status,
		&o.CreatedAt, &o.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return o, nil
}

type TripOfferRepository struct {
	pool *pgxpool.Pool
}

func NewTripOfferRepository(pool *pgxpool.Pool) *TripOfferRepository {
	return &TripOfferRepository{pool: pool}
}

func (r *TripOfferRepository) Create(ctx context.Context, offer *entities.TripOffer) error {
	query := `INSERT INTO trip_offers (` + tripOfferColumns + `)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`

	_, err := r.pool.Exec(ctx, query,
		offer.ID, offer.TripID, offer.DriverID, offer.OfferType, offer.OfferedFare, offer.Status,
		offer.CreatedAt, offer.UpdatedAt,
	)
	return err
}

func (r *TripOfferRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.TripOffer, error) {
	query := `SELECT ` + tripOfferColumns + ` FROM trip_offers WHERE id = $1`

	offer, err := scanTripOffer(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return offer, nil
}

func (r *TripOfferRepository) FindByTripID(ctx context.Context, tripID uuid.UUID) ([]*entities.TripOffer, error) {
	query := `SELECT ` + tripOfferColumns + ` FROM trip_offers WHERE trip_id = $1 ORDER BY offered_fare ASC`

	rows, err := r.pool.Query(ctx, query, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var offers []*entities.TripOffer
	for rows.Next() {
		offer, err := scanTripOffer(rows)
		if err != nil {
			return nil, err
		}
		offers = append(offers, offer)
	}
	return offers, rows.Err()
}

func (r *TripOfferRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entities.OfferStatus) error {
	query := `UPDATE trip_offers SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, status, id)
	return err
}

func (r *TripOfferRepository) RejectOthersForTrip(ctx context.Context, tripID, exceptOfferID uuid.UUID) error {
	query := `UPDATE trip_offers SET status = $1, updated_at = NOW()
		WHERE trip_id = $2 AND id != $3 AND status = $4`
	_, err := r.pool.Exec(ctx, query,
		entities.OfferStatusRejected, tripID, exceptOfferID, entities.OfferStatusPending,
	)
	return err
}

func (r *TripOfferRepository) ExpireAllForTrip(ctx context.Context, tripID uuid.UUID) error {
	query := `UPDATE trip_offers SET status = $1, updated_at = NOW()
		WHERE trip_id = $2 AND status = $3`

	_, err := r.pool.Exec(ctx, query,
		entities.OfferStatusExpired, tripID, entities.OfferStatusPending,
	)
	return err
}
