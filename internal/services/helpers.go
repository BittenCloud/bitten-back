package services

import (
	"bitback/internal/models/customTypes"
	"errors"
	"fmt"
	"time"
)

// calculateEndDate calculates the subscription end date.
func calculateEndDate(startDate time.Time, unit customTypes.DurationUnit, value int) (time.Time, error) {
	if value <= 0 {
		return time.Time{}, errors.New("duration value must be positive")
	}
	switch unit {
	case customTypes.UnitDay:
		return startDate.AddDate(0, 0, value), nil
	case customTypes.UnitMonth:
		return startDate.AddDate(0, value, 0), nil
	case customTypes.UnitYear:
		return startDate.AddDate(value, 0, 0), nil
	default:
		return time.Time{}, fmt.Errorf("invalid duration unit: %s", unit)
	}
}
