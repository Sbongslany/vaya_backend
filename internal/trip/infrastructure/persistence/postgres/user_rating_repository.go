package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

type UserRatingRepository struct {
	pool *pgxpool.Pool
}

func NewUserRatingRepository(pool *pgxpool.Pool) *UserRatingRepository {
	return &UserRatingRepository{pool: pool}
}

func (r *UserRatingRepository) AddRating(ctx context.Context, userID uuid.UUID, rating int) error {
	query := `
		UPDATE auth.users
		SET rating_sum = rating_sum + $1,
			rating_count = rating_count + 1,
			rating_avg = ROUND((rating_sum + $1)::NUMERIC / (rating_count + 1), 2)
		WHERE id = $2
	`
	_, err := r.pool.Exec(ctx, query, rating, userID)
	return err
}

func (r *UserRatingRepository) GetRatingSummary(ctx context.Context, userID uuid.UUID) (*entities.RatingSummary, error) {
	query := `SELECT id, rating_avg, rating_count FROM auth.users WHERE id = $1`

	summary := &entities.RatingSummary{}
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&summary.UserID, &summary.RatingAvg, &summary.RatingCount,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return summary, nil
}
