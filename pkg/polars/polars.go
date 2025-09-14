package polars

import "math"

// Polars interface defines how to get boat speed based on wind conditions
type Polars interface {
	GetBoatSpeed(twa, tws float64) float64
}

// RealisticPolar provides a polar implementation based on actual boat performance data
type RealisticPolar struct{}

// GetBoatSpeed returns boat speed in knots based on TWA (degrees) and TWS (knots)
func (rp *RealisticPolar) GetBoatSpeed(twa, tws float64) float64 {
	// Normalize TWA to 0-180 degrees (absolute angle)
	absTWA := math.Abs(twa)
	if absTWA > 180 {
		absTWA = 360 - absTWA
	}

	// Can't sail closer than 30 degrees to wind
	if absTWA < 30 {
		return 0 // In irons
	}

	// Wind speed data points in the table
	windSpeeds := []float64{4, 6, 8, 10, 12, 14, 16, 20, 24}

	// Angle data points and corresponding speeds for each wind speed
	angles := []float64{52, 60, 75, 90, 110, 120, 135, 150}

	// Speed table: [wind_speed_index][angle_index]
	speedTable := [][]float64{
		{3.73, 3.94, 4.06, 3.99, 4.02, 3.85, 3.37, 2.78},   // 4 kt wind
		{5.05, 5.30, 5.45, 5.47, 5.53, 5.34, 4.77, 4.03},   // 6 kt wind
		{6.01, 6.25, 6.41, 6.55, 6.64, 6.49, 5.96, 5.18},   // 8 kt wind
		{6.62, 6.79, 6.93, 7.13, 7.23, 7.14, 6.81, 6.16},   // 10 kt wind
		{6.94, 7.10, 7.24, 7.47, 7.65, 7.58, 7.33, 6.89},   // 12 kt wind
		{7.08, 7.27, 7.48, 7.67, 8.04, 7.99, 7.76, 7.35},   // 14 kt wind
		{7.16, 7.35, 7.65, 7.82, 8.39, 8.41, 8.21, 7.76},   // 16 kt wind
		{7.24, 7.46, 7.84, 8.19, 8.89, 9.34, 9.24, 8.64},   // 20 kt wind
		{7.26, 7.49, 7.95, 8.44, 9.30, 10.14, 10.85, 9.90}, // 24 kt wind
	}

	// Handle close-hauled angles (30-52 degrees) using beat VMG
	beatVMG := []float64{2.40, 3.33, 4.09, 4.63, 4.96, 5.10, 5.17, 5.24, 5.20}
	beatAngles := []float64{42.7, 42.7, 40.4, 38.9, 37.5, 36.9, 36.6, 36.6, 37.2}

	if absTWA < 52 {
		// Find wind speed index
		windIndex := rp.findWindIndex(tws, windSpeeds)
		beatAngle := rp.interpolateFloat(tws, windSpeeds, beatAngles, windIndex)

		if absTWA < beatAngle {
			return 0 // Can't sail this close
		}

		// Interpolate between no-go zone and 52 degree speed
		vmg := rp.interpolateFloat(tws, windSpeeds, beatVMG, windIndex)
		speed52 := rp.getSpeedAtAngle(52, tws, windSpeeds, angles, speedTable)

		// Linear interpolation between beat angle and 52 degrees
		factor := (absTWA - beatAngle) / (52 - beatAngle)
		return vmg + (speed52-vmg)*factor
	}

	// For angles 52-150, use the speed table
	return rp.getSpeedAtAngle(absTWA, tws, windSpeeds, angles, speedTable)
}

// Helper function to find the appropriate wind speed index
func (rp *RealisticPolar) findWindIndex(tws float64, windSpeeds []float64) int {
	for i := 0; i < len(windSpeeds)-1; i++ {
		if tws <= windSpeeds[i+1] {
			return i
		}
	}
	return len(windSpeeds) - 2 // Return second to last index for extrapolation
}

// Helper function to interpolate a float value from an array
func (rp *RealisticPolar) interpolateFloat(tws float64, windSpeeds []float64, values []float64, windIndex int) float64 {
	if windIndex >= len(windSpeeds)-1 {
		return values[len(values)-1]
	}

	w1, w2 := windSpeeds[windIndex], windSpeeds[windIndex+1]
	v1, v2 := values[windIndex], values[windIndex+1]

	factor := (tws - w1) / (w2 - w1)
	return v1 + (v2-v1)*factor
}

// Helper function to get speed at a specific angle
func (rp *RealisticPolar) getSpeedAtAngle(twa, tws float64, windSpeeds, angles []float64, speedTable [][]float64) float64 {
	windIndex := rp.findWindIndex(tws, windSpeeds)

	// Find angle index
	angleIndex := 0
	for i := 0; i < len(angles)-1; i++ {
		if twa <= angles[i+1] {
			angleIndex = i
			break
		}
	}

	if angleIndex >= len(angles)-1 {
		angleIndex = len(angles) - 2
	}

	// Interpolate between wind speeds
	speed1 := rp.interpolateAngle(twa, angles, speedTable[windIndex], angleIndex)
	speed2 := rp.interpolateAngle(twa, angles, speedTable[windIndex+1], angleIndex)

	// Interpolate between the two wind speed results
	w1, w2 := windSpeeds[windIndex], windSpeeds[windIndex+1]
	factor := (tws - w1) / (w2 - w1)

	return speed1 + (speed2-speed1)*factor
}

// Helper function to interpolate between angles
func (rp *RealisticPolar) interpolateAngle(twa float64, angles, speeds []float64, angleIndex int) float64 {
	if angleIndex >= len(angles)-1 {
		return speeds[len(speeds)-1]
	}

	a1, a2 := angles[angleIndex], angles[angleIndex+1]
	s1, s2 := speeds[angleIndex], speeds[angleIndex+1]

	factor := (twa - a1) / (a2 - a1)
	return s1 + (s2-s1)*factor
}
