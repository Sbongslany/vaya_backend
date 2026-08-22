package usecases

import (
	"context"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type CalculateRouteInput struct {
	FromLat float64
	FromLng float64
	ToLat   float64
	ToLng   float64
}

type CalculateRoute struct {
	routingService services.RoutingService
}

func NewCalculateRoute(routingService services.RoutingService) *CalculateRoute {
	return &CalculateRoute{routingService: routingService}
}

func (uc *CalculateRoute) Execute(ctx context.Context, input CalculateRouteInput) (*services.RouteResult, error) {
	return uc.routingService.CalculateRoute(ctx, input.FromLat, input.FromLng, input.ToLat, input.ToLng)
}
