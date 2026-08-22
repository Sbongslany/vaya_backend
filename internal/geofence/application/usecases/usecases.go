package usecases

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/yourorg/ehailing/backend/internal/geofence/domain/entities"
	"github.com/yourorg/ehailing/backend/internal/geofence/domain/repositories"
)

// --- Create Geofence ---

type CreateGeofenceInput struct {
	Name        string
	Type        entities.GeofenceType
	Coordinates string
	IsActive    bool
}

type CreateGeofence struct {
	geofenceRepo repositories.GeofenceRepository
}

func NewCreateGeofence(geofenceRepo repositories.GeofenceRepository) *CreateGeofence {
	return &CreateGeofence{geofenceRepo: geofenceRepo}
}

func (uc *CreateGeofence) Execute(ctx context.Context, input CreateGeofenceInput) (*entities.Geofence, error) {
	now := time.Now()
	fence := &entities.Geofence{
		ID:          uuid.New(),
		Name:        input.Name,
		Type:        input.Type,
		Coordinates: input.Coordinates,
		IsActive:    input.IsActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.geofenceRepo.Create(ctx, fence); err != nil {
		return nil, err
	}
	return fence, nil
}

// --- List Geofences ---

type ListGeofences struct {
	geofenceRepo repositories.GeofenceRepository
}

func NewListGeofences(geofenceRepo repositories.GeofenceRepository) *ListGeofences {
	return &ListGeofences{geofenceRepo: geofenceRepo}
}

func (uc *ListGeofences) Execute(ctx context.Context, activeOnly bool) ([]*entities.Geofence, error) {
	return uc.geofenceRepo.List(ctx, activeOnly)
}

// --- Check Location In Geofence ---

type CheckLocationInGeofence struct {
	geofenceRepo repositories.GeofenceRepository
}

func NewCheckLocationInGeofence(geofenceRepo repositories.GeofenceRepository) *CheckLocationInGeofence {
	return &CheckLocationInGeofence{geofenceRepo: geofenceRepo}
}

func (uc *CheckLocationInGeofence) Execute(ctx context.Context, lat, lng float64) ([]*entities.Geofence, error) {
	return uc.geofenceRepo.FindZonesContainingPoint(ctx, lat, lng)
}

// --- Assign Driver To Zone ---

type AssignDriverToZoneInput struct {
	DriverID uuid.UUID
	ZoneID   uuid.UUID
	Status   string
}

type AssignDriverToZone struct {
	assignmentRepo repositories.ZoneAssignmentRepository
}

func NewAssignDriverToZone(assignmentRepo repositories.ZoneAssignmentRepository) *AssignDriverToZone {
	return &AssignDriverToZone{assignmentRepo: assignmentRepo}
}

func (uc *AssignDriverToZone) Execute(ctx context.Context, input AssignDriverToZoneInput) error {
	assignment := &entities.ZoneAssignment{
		ID:        uuid.New(),
		DriverID:  input.DriverID,
		ZoneID:    input.ZoneID,
		Status:    input.Status,
		CreatedAt: time.Now(),
	}
	return uc.assignmentRepo.AssignDriver(ctx, assignment)
}

// --- Remove Driver From Zone ---

type RemoveDriverFromZone struct {
	assignmentRepo repositories.ZoneAssignmentRepository
}

func NewRemoveDriverFromZone(assignmentRepo repositories.ZoneAssignmentRepository) *RemoveDriverFromZone {
	return &RemoveDriverFromZone{assignmentRepo: assignmentRepo}
}

func (uc *RemoveDriverFromZone) Execute(ctx context.Context, driverID, zoneID uuid.UUID) error {
	return uc.assignmentRepo.RemoveDriver(ctx, driverID, zoneID)
}
