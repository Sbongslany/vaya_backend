package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/support/domain/entities"
)

type TicketRepository interface {
	Create(ctx context.Context, ticket *entities.SupportTicket) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.SupportTicket, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.SupportTicket, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entities.TicketStatus, adminID *uuid.UUID) error
}

type CommentRepository interface {
	Create(ctx context.Context, comment *entities.TicketComment) error
	FindByTicketID(ctx context.Context, ticketID uuid.UUID) ([]*entities.TicketComment, error)
}

type RefundRepository interface {
	Create(ctx context.Context, refund *entities.Refund) error
	FindByTicketID(ctx context.Context, ticketID uuid.UUID) (*entities.Refund, error)
}
