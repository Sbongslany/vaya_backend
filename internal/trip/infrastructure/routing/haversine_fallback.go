package routing

import (
	"context"
	"math"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/services"
)

type HaversineFallback struct{}

func NewHaversineFallback() *HaversineFallback {
	return &HaversineFallback{}
}

func (f *HaversineFallback) CalculateRoute(ctx context.Context, fromLat, fromLng, toLat, toLng float64) (*services.RouteResult, error) {
	distanceKM := calculateHaversineDistance(fromLat, fromLng, toLat, toLng)

	// Add 30% to account for road curvature vs straight-line
	roadDistance := distanceKM * 1.3

	// Estimate duration: assume average city speed of 30 km/h
	durationMinutes := int((roadDistance / 30.0) * 60.0)

	return &services.RouteResult{
		DistanceKM:      math.Round(roadDistance*100) / 100,
		DurationMinutes: durationMinutes,
		Polyline:        "", // No polyline for Haversine fallback
		Steps:           nil,
	}, nil
}

func calculateHaversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusKM = 6371.0

	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLng := (lng2 - lng1) * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKM * c
}
