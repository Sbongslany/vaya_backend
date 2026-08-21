package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/support/domain"
	"github.com/yourorg/ehailing/backend/internal/support/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/support/domain/repositories"
	"github.com/yourorg/ehailing/backend/internal/support/domain/services"
)

// --- Create Ticket ---

type CreateTicketInput struct {
	UserID      uuid.UUID
	TripID      *uuid.UUID
	Category    entities.TicketCategory
	Subject     string
	Description string
}

type CreateTicket struct {
	ticketRepo repositories.TicketRepository
}

func NewCreateTicket(ticketRepo repositories.TicketRepository) *CreateTicket {
	return &CreateTicket{ticketRepo: ticketRepo}
}

func (uc *CreateTicket) Execute(ctx context.Context, input CreateTicketInput) (*entities.SupportTicket, error) {
	now := time.Now()
	ticket := &entities.SupportTicket{
		ID:          uuid.New(),
		UserID:      input.UserID,
		TripID:      input.TripID,
		Category:    input.Category,
		Subject:     input.Subject,
		Description: input.Description,
		Status:      entities.StatusOpen,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.ticketRepo.Create(ctx, ticket); err != nil {
		return nil, err
	}
	return ticket, nil
}

// --- Get User Tickets ---

type GetUserTickets struct {
	ticketRepo repositories.TicketRepository
}

func NewGetUserTickets(ticketRepo repositories.TicketRepository) *GetUserTickets {
	return &GetUserTickets{ticketRepo: ticketRepo}
}

func (uc *GetUserTickets) Execute(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.SupportTicket, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return uc.ticketRepo.FindByUserID(ctx, userID, limit, offset)
}

// --- Add Comment ---

type AddCommentInput struct {
	TicketID uuid.UUID
	AuthorID uuid.UUID
	Content  string
}

type AddComment struct {
	ticketRepo  repositories.TicketRepository
	commentRepo repositories.CommentRepository
}

func NewAddComment(ticketRepo repositories.TicketRepository, commentRepo repositories.CommentRepository) *AddComment {
	return &AddComment{ticketRepo: ticketRepo, commentRepo: commentRepo}
}

func (uc *AddComment) Execute(ctx context.Context, input AddCommentInput) (*entities.TicketComment, error) {
	ticket, err := uc.ticketRepo.GetByID(ctx, input.TicketID)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, domain.ErrTicketNotFound
	}

	// Ensure the author is either the ticket creator or an assigned admin
	if ticket.UserID != input.AuthorID && (ticket.AdminAssigned == nil || *ticket.AdminAssigned != input.AuthorID) {
		return nil, domain.ErrUnauthorized
	}

	comment := &entities.TicketComment{
		ID:        uuid.New(),
		TicketID:  input.TicketID,
		AuthorID:  input.AuthorID,
		Content:   input.Content,
		CreatedAt: time.Now(),
	}

	if err := uc.commentRepo.Create(ctx, comment); err != nil {
		return nil, err
	}

	// Update ticket to IN_PROGRESS if it was OPEN and admin replied
	if ticket.Status == entities.StatusOpen && ticket.UserID != input.AuthorID {
		_ = uc.ticketRepo.UpdateStatus(ctx, ticket.ID, entities.StatusInProgress, &input.AuthorID)
	}

	return comment, nil
}

// --- Admin Resolve Ticket (with optional refund) ---

type ResolveTicketInput struct {
	TicketID     uuid.UUID
	AdminID      uuid.UUID
	Resolution   string  // Admin's final comment
	RefundAmount float64 // 0 means no refund
	RefundReason string
}

type ResolveTicket struct {
	ticketRepo     repositories.TicketRepository
	commentRepo    repositories.CommentRepository
	refundRepo     repositories.RefundRepository
	walletCreditor services.WalletCreditor
}

func NewResolveTicket(
	ticketRepo repositories.TicketRepository,
	commentRepo repositories.CommentRepository,
	refundRepo repositories.RefundRepository,
	walletCreditor services.WalletCreditor,
) *ResolveTicket {
	return &ResolveTicket{
		ticketRepo:     ticketRepo,
		commentRepo:    commentRepo,
		refundRepo:     refundRepo,
		walletCreditor: walletCreditor,
	}
}

func (uc *ResolveTicket) Execute(ctx context.Context, input ResolveTicketInput) error {
	ticket, err := uc.ticketRepo.GetByID(ctx, input.TicketID)
	if err != nil {
		return err
	}
	if ticket == nil {
		return domain.ErrTicketNotFound
	}

	if ticket.Status == entities.StatusResolved || ticket.Status == entities.StatusClosed {
		return domain.ErrInvalidStatus
	}

	// Add resolution comment
	resComment := &entities.TicketComment{
		ID:        uuid.New(),
		TicketID:  input.TicketID,
		AuthorID:  input.AdminID,
		Content:   "Resolution: " + input.Resolution,
		CreatedAt: time.Now(),
	}
	if err := uc.commentRepo.Create(ctx, resComment); err != nil {
		return err
	}

	// Process refund if amount > 0
	if input.RefundAmount > 0 {
		if uc.walletCreditor == nil {
			return domain.ErrInvalidAmount // Should not happen if wired correctly
		}

		err := uc.walletCreditor.CreditUserWallet(ctx, ticket.UserID, input.RefundAmount, input.RefundReason, &input.AdminID)
		if err != nil {
			return err
		}

		refund := &entities.Refund{
			ID:          uuid.New(),
			TicketID:    input.TicketID,
			UserID:      ticket.UserID,
			Amount:      input.RefundAmount,
			Reason:      input.RefundReason,
			ProcessedBy: &input.AdminID,
			CreatedAt:   time.Now(),
		}
		if err := uc.refundRepo.Create(ctx, refund); err != nil {
			return err
		}
	}

	// Mark ticket as resolved
	return uc.ticketRepo.UpdateStatus(ctx, ticket.ID, entities.StatusResolved, &input.AdminID)
}
