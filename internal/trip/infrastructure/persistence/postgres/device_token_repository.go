package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

type DeviceTokenRepository struct {
	pool *pgxpool.Pool
}

func NewDeviceTokenRepository(pool *pgxpool.Pool) *DeviceTokenRepository {
	return &DeviceTokenRepository{pool: pool}
}

func (r *DeviceTokenRepository) Save(ctx context.Context, token *entities.DeviceToken) error {
	query := `
		INSERT INTO device_tokens (user_id, token, device_type)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, token) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, token.UserID, token.Token, token.DeviceType)
	return err
}

func (r *DeviceTokenRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.DeviceToken, error) {
	query := `SELECT user_id, token, device_type, created_at FROM device_tokens WHERE user_id = $1`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*entities.DeviceToken
	for rows.Next() {
		t := &entities.DeviceToken{}
		if err := rows.Scan(&t.UserID, &t.Token, &t.DeviceType, &t.CreatedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}

func (r *DeviceTokenRepository) Delete(ctx context.Context, userID uuid.UUID, token string) error {
	query := `DELETE FROM device_tokens WHERE user_id = $1 AND token = $2`
	_, err := r.pool.Exec(ctx, query, userID, token)
	return err
}
