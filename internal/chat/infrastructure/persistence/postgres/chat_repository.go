package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/chat/domain/entities"
)

const chatColumns = `id, trip_id, sender_id, receiver_id, content, read_at, created_at`

type ChatRepository struct {
	pool *pgxpool.Pool
}

func NewChatRepository(pool *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{pool: pool}
}

func (r *ChatRepository) Create(ctx context.Context, message *entities.ChatMessage) error {
	query := `INSERT INTO chat_messages (` + chatColumns + `)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		message.ID, message.TripID, message.SenderID, message.ReceiverID,
		message.Content, message.ReadAt, message.CreatedAt,
	)
	return err
}

func (r *ChatRepository) FindByTripID(ctx context.Context, tripID uuid.UUID, limit, offset int) ([]*entities.ChatMessage, error) {
	query := `SELECT ` + chatColumns + ` FROM chat_messages
		WHERE trip_id = $1 ORDER BY created_at ASC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, tripID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*entities.ChatMessage
	for rows.Next() {
		m := &entities.ChatMessage{}
		if err := rows.Scan(&m.ID, &m.TripID, &m.SenderID, &m.ReceiverID, &m.Content, &m.ReadAt, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

func (r *ChatRepository) MarkAsRead(ctx context.Context, tripID uuid.UUID, readerID uuid.UUID) error {
	query := `UPDATE chat_messages SET read_at = NOW()
		WHERE trip_id = $1 AND receiver_id = $2 AND read_at IS NULL`
	_, err := r.pool.Exec(ctx, query, tripID, readerID)
	return err
}
