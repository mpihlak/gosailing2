package world

import (
	"math"
	"math/rand"
	"time"

	"github.com/mpihlak/gosailing2/pkg/geometry"
)

type Wind interface {
	GetWind(pos geometry.Point) (direction, speed float64)
}

type ConstantWind struct {
	Direction float64
	Speed     float64
}

func (cw *ConstantWind) GetWind(_ geometry.Point) (float64, float64) {
	return cw.Direction, cw.Speed
}

// VariableWind provides wind that varies in strength across the course
type VariableWind struct {
	Direction  float64 // Wind direction (constant)
	LeftSpeed  float64 // Wind speed on left side (X=0)
	RightSpeed float64 // Wind speed on right side (X=WorldWidth)
	WorldWidth float64 // Width of the world for interpolation
}

func (vw *VariableWind) GetWind(pos geometry.Point) (float64, float64) {
	// Validate inputs to prevent NaN
	if math.IsNaN(pos.X) || math.IsInf(pos.X, 0) || vw.WorldWidth <= 0 {
		return vw.Direction, vw.LeftSpeed // Return safe fallback
	}

	// Interpolate wind speed based on X position
	// X=0 (left) = LeftSpeed, X=WorldWidth (right) = RightSpeed
	xRatio := pos.X / vw.WorldWidth
	if xRatio < 0 {
		xRatio = 0
	} else if xRatio > 1 {
		xRatio = 1
	}

	// Linear interpolation between left and right speeds
	speed := vw.LeftSpeed + (vw.RightSpeed-vw.LeftSpeed)*xRatio

	// Validate result
	if math.IsNaN(speed) || math.IsInf(speed, 0) || speed < 0 {
		speed = vw.LeftSpeed // Fallback to left speed
	}

	return vw.Direction, speed
}

// OscillatingWind wraps VariableWind with random directional oscillations
type OscillatingWind struct {
	baseWind        *VariableWind
	medianDirection float64 // Base wind direction (0 = North)

	// Oscillation state
	shiftStartTime   time.Time     // When current shift started
	shiftDuration    time.Duration // How long this shift lasts
	shiftAngle       float64       // Target shift angle (-10 to +10 degrees)
	currentDirection float64       // Current wind direction including shift

	// Phase tracking
	shiftPhase     int           // 0=shifting out, 1=at peak, 2=shifting back
	phaseStartTime time.Time     // When current phase started
	phaseDuration  time.Duration // Duration of current phase

	// Start line bias (initial oscillation)
	isInitialBias      bool      // Whether this is the first bias oscillation
	initialBiasAngle   float64   // Fixed bias angle for initial oscillation
	gameStartTime      time.Time // When the game started (for 3s delay)
	isInInitialBiasCycle bool    // Whether we're currently executing the initial bias cycle
}

func NewOscillatingWind(leftSpeed, rightSpeed, worldWidth float64) *OscillatingWind {
	// Randomly determine start line bias
	// Positive angle = committee boat favored (starboard tack lift)
	// Negative angle = pin favored (port tack lift)
	biasDirection := 1.0
	if rand.Float32() < 0.5 {
		biasDirection = -1.0 // Pin favored
	}
	// Random bias between 5 and 15 degrees
	biasAngle := biasDirection * (5.0 + rand.Float64()*10.0)

	now := time.Now()
	ow := &OscillatingWind{
		baseWind: &VariableWind{
			Direction:  0,
			LeftSpeed:  leftSpeed,
			RightSpeed: rightSpeed,
			WorldWidth: worldWidth,
		},
		medianDirection:  0, // North
		currentDirection: 0,
		shiftPhase:       0,
		shiftStartTime:   now,
		phaseStartTime:   now,
		// Initial bias setup
		isInitialBias:    true,
		initialBiasAngle: biasAngle,
		gameStartTime:    now,
	}
	// Initialize first shift with bias - will start after 3 second delay
	ow.startNewShift(now)
	return ow
}

func (ow *OscillatingWind) Update() {
	ow.UpdateWithElapsedTime(0)
}

func (ow *OscillatingWind) UpdateWithElapsedTime(gameElapsedSeconds float64) {
	now := time.Now()

	// Check if we need to start a new shift cycle
	if ow.shiftPhase == 0 && ow.shiftStartTime.IsZero() {
		// Initialize first shift
		ow.startNewShift(now)
	}

	elapsedPhase := now.Sub(ow.phaseStartTime)

	switch ow.shiftPhase {
	case 0: // Shifting out from median to target angle
		progress := float64(elapsedPhase) / float64(ow.phaseDuration)
		if progress >= 1.0 {
			// Phase complete, move to peak
			ow.shiftPhase = 1
			ow.phaseStartTime = now
			if ow.isInInitialBiasCycle {
				// Phase 1 (peak): 25 seconds (10s-35s)
				ow.phaseDuration = 25 * time.Second
			} else {
				ow.phaseDuration = ow.shiftDuration / 3 // Peak lasts 1/3 of total duration
			}
			ow.currentDirection = ow.medianDirection + ow.shiftAngle
		} else {
			// Interpolate toward target angle
			ow.currentDirection = ow.medianDirection + ow.shiftAngle*progress
		}

	case 1: // At peak angle
		if elapsedPhase >= ow.phaseDuration {
			// Peak complete, start shifting back
			ow.shiftPhase = 2
			ow.phaseStartTime = now
			if ow.isInInitialBiasCycle {
				// Phase 2 (back): 10 seconds (35s-45s)
				ow.phaseDuration = 10 * time.Second
			} else {
				ow.phaseDuration = ow.shiftDuration / 3 // Return phase lasts 1/3 of total duration
			}
		}
		// Stay at peak angle
		ow.currentDirection = ow.medianDirection + ow.shiftAngle

	case 2: // Shifting back from target angle to median
		progress := float64(elapsedPhase) / float64(ow.phaseDuration)
		if progress >= 1.0 {
			// Shift cycle complete, start new one
			ow.currentDirection = ow.medianDirection
			ow.startNewShift(now)
		} else {
			// Interpolate back to median
			ow.currentDirection = ow.medianDirection + ow.shiftAngle*(1.0-progress)
		}
	}

	// Normalize direction to 0-360 range
	for ow.currentDirection < 0 {
		ow.currentDirection += 360
	}
	for ow.currentDirection >= 360 {
		ow.currentDirection -= 360
	}

	// Update the base wind direction
	ow.baseWind.Direction = ow.currentDirection
}

func (ow *OscillatingWind) startNewShift(now time.Time) {
	// Check if this is the initial bias shift
	if ow.isInitialBias {
		// Use fixed bias parameters for initial shift
		// Timeline: shift to bias by ~10s, hold until 35s, revert by ~45s
		ow.shiftDuration = 45 * time.Second // Total duration: 10s out + 25s peak + 10s back
		ow.shiftAngle = ow.initialBiasAngle // Use predetermined bias angle
		ow.isInInitialBiasCycle = true      // Mark that we're in the initial bias cycle

		// Mark that we've started the initial bias
		// After this shift completes, we'll transition to normal oscillations
		ow.isInitialBias = false
	} else {
		// Normal random shift parameters
		ow.shiftDuration = time.Duration(13+rand.Intn(13)) * time.Second // 13-25 seconds
		ow.shiftAngle = -10.0 + rand.Float64()*20.0                      // -10 to +10 degrees
		ow.isInInitialBiasCycle = false
	}

	// Reset shift state
	ow.shiftPhase = 0
	ow.shiftStartTime = now
	ow.phaseStartTime = now

	if ow.isInInitialBiasCycle {
		// Initial bias has custom phase durations
		// Phase 0 (out): 10 seconds (0s-10s)
		ow.phaseDuration = 10 * time.Second
	} else {
		// Normal oscillation: 1/3 for each phase
		ow.phaseDuration = ow.shiftDuration / 3
	}
}

func (ow *OscillatingWind) GetWind(pos geometry.Point) (float64, float64) {
	return ow.baseWind.GetWind(pos)
}
