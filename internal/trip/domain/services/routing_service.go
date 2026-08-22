package services

import "context"

type RouteStep struct {
	Instruction     string
	DistanceKM      float64
	DurationMinutes int
}

type RouteResult struct {
	DistanceKM      float64
	DurationMinutes int
	Polyline        string
	Steps           []RouteStep
}

type RoutingService interface {
	CalculateRoute(ctx context.Context, fromLat, fromLng, toLat, toLng float64) (*RouteResult, error)
}
