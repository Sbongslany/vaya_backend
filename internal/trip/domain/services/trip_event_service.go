package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/trip/domain/repositories"
)

type TripEventService struct {
	eventRepo   repositories.TripEventRepository
	broadcaster EventBroadcaster
}

func NewTripEventService(eventRepo repositories.TripEventRepository, broadcaster EventBroadcaster) *TripEventService {
	return &TripEventService{eventRepo: eventRepo, broadcaster: broadcaster}
}

func (s *TripEventService) Record(
	ctx context.Context,
	tripID uuid.UUID,
	eventType string,
	actorID *uuid.UUID,
	fromStatus, toStatus string,
	metadata map[string]interface{},
) error {
	var metadataJSON json.RawMessage
	if metadata != nil {
		data, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		metadataJSON = data
	}

	event := &entities.TripEvent{
		ID:         uuid.New(),
		TripID:     tripID,
		EventType:  eventType,
		ActorID:    actorID,
		FromStatus: &fromStatus,
		ToStatus:   &toStatus,
		Metadata:   metadataJSON,
		CreatedAt:  time.Now(),
	}

	if err := s.eventRepo.Create(ctx, event); err != nil {
		return err
	}

	// Broadcast to WebSocket clients in real-time
	if s.broadcaster != nil {
		s.broadcaster.Broadcast(tripID, event)
	}

	return nil
}
