package services

import (
	"math"

	"github.com/yourorg/ehailing/backend/internal/trip/domain/entities"
)

type FareCalculator struct {
	BaseFare            float64
	PerKMFare           float64
	PerMinuteFare       float64
	MinimumFare         float64
	RetentionPerDayFare float64
}

func NewFareCalculator() *FareCalculator {
	return &FareCalculator{
		BaseFare:            10.0,
		PerKMFare:           5.0,
		PerMinuteFare:       1.0,
		MinimumFare:         25.0,
		RetentionPerDayFare: 500.0,
	}
}

func (fc *FareCalculator) Calculate(distanceKM float64, durationMinutes int) float64 {
	fare := fc.BaseFare + (distanceKM * fc.PerKMFare) + (float64(durationMinutes) * fc.PerMinuteFare)
	return math.Max(fare, fc.MinimumFare)
}

// CalculateLongDistanceFare computes fare for long-distance trips.
// ONE_WAY: outbound driving fare only.
// RETURN: outbound + return driving fare.
// MULTI_DAY: outbound fare + driver retention per day.
func (fc *FareCalculator) CalculateLongDistanceFare(distanceKM float64, durationMinutes int, ldType entities.LongDistanceType, durationDays int) float64 {
	outboundFare := fc.BaseFare + (distanceKM * fc.PerKMFare) + (float64(durationMinutes) * fc.PerMinuteFare)
	fare := outboundFare

	switch ldType {
	case entities.LongDistanceReturn:
		returnFare := fc.BaseFare + (distanceKM * fc.PerKMFare) + (float64(durationMinutes) * fc.PerMinuteFare)
		fare += returnFare
	case entities.LongDistanceMultiDay:
		if durationDays > 0 {
			fare += float64(durationDays) * fc.RetentionPerDayFare
		}
	}

	return math.Max(fare, fc.MinimumFare)
}

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
