package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/support/domain/entities"
)

// --- Ticket Repository ---

type TicketRepository struct {
	pool *pgxpool.Pool
}

func NewTicketRepository(pool *pgxpool.Pool) *TicketRepository {
	return &TicketRepository{pool: pool}
}

func (r *TicketRepository) Create(ctx context.Context, ticket *entities.SupportTicket) error {
	query := `INSERT INTO support_tickets (id, user_id, trip_id, category, subject, description, status, admin_assigned, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := r.pool.Exec(ctx, query,
		ticket.ID, ticket.UserID, ticket.TripID, ticket.Category, ticket.Subject, ticket.Description,
		ticket.Status, ticket.AdminAssigned, ticket.CreatedAt, ticket.UpdatedAt,
	)
	return err
}

func (r *TicketRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.SupportTicket, error) {
	query := `SELECT id, user_id, trip_id, category, subject, description, status, admin_assigned, created_at, updated_at
		FROM support_tickets WHERE id = $1`

	t := &entities.SupportTicket{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.UserID, &t.TripID, &t.Category, &t.Subject, &t.Description,
		&t.Status, &t.AdminAssigned, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return t, nil
}

func (r *TicketRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.SupportTicket, error) {
	query := `SELECT id, user_id, trip_id, category, subject, description, status, admin_assigned, created_at, updated_at
		FROM support_tickets WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*entities.SupportTicket
	for rows.Next() {
		t := &entities.SupportTicket{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.TripID, &t.Category, &t.Subject, &t.Description, &t.Status, &t.AdminAssigned, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}
	return tickets, rows.Err()
}

func (r *TicketRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entities.TicketStatus, adminID *uuid.UUID) error {
	query := `UPDATE support_tickets SET status = $1, admin_assigned = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, status, adminID, id)
	return err
}

// --- Comment Repository ---

type CommentRepository struct {
	pool *pgxpool.Pool
}

func NewCommentRepository(pool *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{pool: pool}
}

func (r *CommentRepository) Create(ctx context.Context, comment *entities.TicketComment) error {
	query := `INSERT INTO ticket_comments (id, ticket_id, author_id, content, created_at)
		VALUES ($1,$2,$3,$4,$5)`
	_, err := r.pool.Exec(ctx, query, comment.ID, comment.TicketID, comment.AuthorID, comment.Content, comment.CreatedAt)
	return err
}

func (r *CommentRepository) FindByTicketID(ctx context.Context, ticketID uuid.UUID) ([]*entities.TicketComment, error) {
	query := `SELECT id, ticket_id, author_id, content, created_at FROM ticket_comments
		WHERE ticket_id = $1 ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []*entities.TicketComment
	for rows.Next() {
		c := &entities.TicketComment{}
		if err := rows.Scan(&c.ID, &c.TicketID, &c.AuthorID, &c.Content, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

// --- Refund Repository ---

type RefundRepository struct {
	pool *pgxpool.Pool
}

func NewRefundRepository(pool *pgxpool.Pool) *RefundRepository {
	return &RefundRepository{pool: pool}
}

func (r *RefundRepository) Create(ctx context.Context, refund *entities.Refund) error {
	query := `INSERT INTO refunds (id, ticket_id, user_id, amount, reason, processed_by, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := r.pool.Exec(ctx, query, refund.ID, refund.TicketID, refund.UserID, refund.Amount, refund.Reason, refund.ProcessedBy, refund.CreatedAt)
	return err
}

func (r *RefundRepository) FindByTicketID(ctx context.Context, ticketID uuid.UUID) (*entities.Refund, error) {
	query := `SELECT id, ticket_id, user_id, amount, reason, processed_by, created_at FROM refunds WHERE ticket_id = $1`
	ref := &entities.Refund{}
	err := r.pool.QueryRow(ctx, query, ticketID).Scan(&ref.ID, &ref.TicketID, &ref.UserID, &ref.Amount, &ref.Reason, &ref.ProcessedBy, &ref.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return ref, nil
}
