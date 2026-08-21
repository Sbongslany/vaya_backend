package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

const ratingColumns = `id, trip_id, rater_id, rated_user_id, rating, comment, created_at`

type TripRatingRepository struct {
	pool *pgxpool.Pool
}

func NewTripRatingRepository(pool *pgxpool.Pool) *TripRatingRepository {
	return &TripRatingRepository{pool: pool}
}

func (r *TripRatingRepository) Create(ctx context.Context, rating *entities.TripRating) error {
	query := `INSERT INTO trip_ratings (` + ratingColumns + `)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		rating.ID, rating.TripID, rating.RaterID, rating.RatedUserID,
		rating.Rating, rating.Comment, rating.CreatedAt,
	)
	return err
}

func (r *TripRatingRepository) FindByTripAndRater(ctx context.Context, tripID uuid.UUID, raterID uuid.UUID) (*entities.TripRating, error) {
	query := `SELECT ` + ratingColumns + ` FROM trip_ratings WHERE trip_id = $1 AND rater_id = $2`

	rating := &entities.TripRating{}
	err := r.pool.QueryRow(ctx, query, tripID, raterID).Scan(
		&rating.ID, &rating.TripID, &rating.RaterID, &rating.RatedUserID,
		&rating.Rating, &rating.Comment, &rating.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return rating, nil
}

func (r *TripRatingRepository) CountByTripID(ctx context.Context, tripID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM trip_ratings WHERE trip_id = $1`
	var count int
	err := r.pool.QueryRow(ctx, query, tripID).Scan(&count)
	return count, err
}
