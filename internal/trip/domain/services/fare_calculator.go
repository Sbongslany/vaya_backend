package services

import "math"

type FareCalculator struct {
	BaseFare      float64
	PerKMFare     float64
	PerMinuteFare float64
	MinimumFare   float64
}

func NewFareCalculator() *FareCalculator {
	return &FareCalculator{
		BaseFare:      10.0,  // R10
		PerKMFare:     5.0,   // R5 per km
		PerMinuteFare: 1.0,   // R1 per min
		MinimumFare:   25.0,  // R25 minimum
	}
}

func (fc *FareCalculator) Calculate(distanceKM float64, durationMinutes int) float64 {
	fare := fc.BaseFare + (distanceKM * fc.PerKMFare) + (float64(durationMinutes) * fc.PerMinuteFare)
	return math.Max(fare, fc.MinimumFare)
}

// Haversine formula for distance
func CalculateDistanceKM(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLng := (lng2 - lng1) * math.Pi / 180.0
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}