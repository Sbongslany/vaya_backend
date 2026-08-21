package services

import "math"

type CommissionService struct {
	Rate float64 // e.g., 0.20 for 20%
}

func NewCommissionService(rate float64) *CommissionService {
	return &CommissionService{Rate: rate}
}

func (s *CommissionService) Calculate(fare float64) (commission float64, driverEarnings float64) {
	commission = math.Round(fare*s.Rate*100) / 100
	driverEarnings = math.Round((fare-commission)*100) / 100
	return commission, driverEarnings
}
