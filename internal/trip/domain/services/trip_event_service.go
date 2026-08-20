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
	eventRepo repositories.TripEventRepository
}

func NewTripEventService(eventRepo repositories.TripEventRepository) *TripEventService {
	return &TripEventService{eventRepo: eventRepo}
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

	return s.eventRepo.Create(ctx, event)
}
