package usecases

import (
	"context"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type GetSurgeMultiplier struct {
	surgeService services.SurgeService
}

func NewGetSurgeMultiplier(surgeService services.SurgeService) *GetSurgeMultiplier {
	return &GetSurgeMultiplier{surgeService: surgeService}
}

func (uc *GetSurgeMultiplier) Execute(ctx context.Context, lat, lng float64) (float64, error) {
	return uc.surgeService.GetMultiplier(ctx, lat, lng)
}

type GetSurgeHeatmap struct {
	surgeService services.SurgeService
}

func NewGetSurgeHeatmap(surgeService services.SurgeService) *GetSurgeHeatmap {
	return &GetSurgeHeatmap{surgeService: surgeService}
}

func (uc *GetSurgeHeatmap) Execute(ctx context.Context) ([]services.SurgeZone, error) {
	return uc.surgeService.GetHeatmap(ctx)
}
