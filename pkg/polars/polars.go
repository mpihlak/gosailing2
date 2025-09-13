package polars

import "math"

// Polars interface defines how to get boat speed based on wind conditions
type Polars interface {
	GetBoatSpeed(twa, tws float64) float64
}

// SimplePolar provides a basic polar implementation for the MVP
type SimplePolar struct{}

// GetBoatSpeed returns boat speed in knots based on TWA (degrees) and TWS (knots)
func (sp *SimplePolar) GetBoatSpeed(twa, tws float64) float64 {
	// Normalize TWA to 0-180 degrees (absolute angle)
	absTWA := math.Abs(twa)
	if absTWA > 180 {
		absTWA = 360 - absTWA
	}

	// Simple polar curve - boat can't sail closer than 45 degrees to wind
	if absTWA < 45 {
		return 0 // In irons - can't sail this close to wind
	}

	// Basic speed calculation based on wind angle
	var speedFactor float64
	switch {
	case absTWA < 60: // Close hauled
		speedFactor = 0.6
	case absTWA < 90: // Close reach
		speedFactor = 0.8
	case absTWA < 120: // Beam reach
		speedFactor = 1.0 // Maximum boat speed
	case absTWA < 150: // Broad reach
		speedFactor = 0.9
	default: // Running
		speedFactor = 0.7
	}

	// Base boat speed is proportional to wind speed, capped at reasonable values
	baseSpeed := tws * 0.6 // Boat speed is typically 60% of wind speed
	if baseSpeed > 8 {     // Cap at 8 knots for this simple model
		baseSpeed = 8
	}

	return baseSpeed * speedFactor
}
