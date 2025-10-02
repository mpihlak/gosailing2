package game

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/mpihlak/gosailing2/pkg/dashboard"
	"github.com/mpihlak/gosailing2/pkg/game/objects"
	"github.com/mpihlak/gosailing2/pkg/game/world"
)

// Telltales represents a single jib telltale that indicates sailing efficiency
type Telltales struct {
	Length  float64 // Length in pixels (75px)
	BaseX   float64 // Screen X position (hinge point)
	BaseY   float64 // Screen Y position (hinge point)
	Angle   float64 // Telltale angle in degrees (0 = horizontal, negative = up, positive = down)
	Visible bool    // Whether telltale should be shown (always true now)
}

// NewTelltales creates a new telltales instance
func NewTelltales(screenWidth, screenHeight int) *Telltales {
	return &Telltales{
		Length:  75.0,
		BaseX:   float64(screenWidth/2 - 50), // Left of center
		BaseY:   80.0,                        // Below timer and OCS warning
		Angle:   0.0,                         // Start horizontal
		Visible: true,                        // Always visible now
	}
}

// Update calculates telltale position based on boat performance
func (t *Telltales) Update(boat *objects.Boat, wind world.Wind, dashboard *dashboard.Dashboard) {
	windDir, windSpeed := wind.GetWind(boat.Pos)
	twa := boat.Heading - windDir

	// Normalize TWA to -180 to +180
	if twa < -180 {
		twa += 360
	} else if twa > 180 {
		twa -= 360
	}

	absTWA := math.Abs(twa)

	// Telltale is always visible in all sailing modes
	t.Visible = true

	// Calculate VMG efficiency
	currentVMG := dashboard.CalculateVMG()
	targetVMG := dashboard.FindBestVMG()

	efficiency := 0.0
	if math.Abs(targetVMG) > 0.001 { // Avoid division by very small numbers
		efficiency = currentVMG / targetVMG
	}

	// Calculate optimal TWA for current wind conditions and sailing mode
	optimalTWA := t.findOptimalTWA(boat, windSpeed, absTWA)

	// Calculate telltale angle based on VMG efficiency
	t.calculateTelltaleAngle(absTWA, optimalTWA, efficiency)
}

// findOptimalTWA finds the optimal TWA for current wind conditions using polars
func (t *Telltales) findOptimalTWA(boat *objects.Boat, windSpeed float64, absTWA float64) float64 {
	bestVMG := 0.0
	bestTWA := 45.0 // Default fallback

	if absTWA <= 90 {
		// Upwind sailing - search for best VMG angle between 30-60 degrees
		for angle := 30.0; angle <= 60.0; angle += 1.0 {
			speed := boat.Polars.GetBoatSpeed(angle, windSpeed)
			angleRad := angle * math.Pi / 180
			vmg := speed * math.Cos(angleRad)

			if vmg > bestVMG {
				bestVMG = vmg
				bestTWA = angle
			}
		}
	} else {
		// Downwind sailing - search for best VMG angle between 120-170 degrees
		bestVMG = 1000.0 // Start with high value for downwind (looking for most negative VMG)
		bestTWA = 150.0  // Default downwind angle
		for angle := 120.0; angle <= 170.0; angle += 1.0 {
			speed := boat.Polars.GetBoatSpeed(angle, windSpeed)
			angleRad := angle * math.Pi / 180
			vmg := speed * math.Cos(angleRad) // This will be negative for downwind

			if vmg < bestVMG { // Most negative VMG is best for downwind
				bestVMG = vmg
				bestTWA = angle
			}
		}
	}

	return bestTWA
}

// calculateTelltaleAngle determines telltale angle based on VMG efficiency
func (t *Telltales) calculateTelltaleAngle(absTWA, optimalTWA, efficiency float64) {
	// Reset angle
	t.Angle = 0.0

	// If sailing very efficiently, keep telltale horizontal
	if efficiency >= 0.95 {
		return
	}

	// Maximum deflection angles
	maxUpAngle := -45.0  // Maximum upward deflection (pinching)
	maxDownAngle := 30.0 // Maximum downward deflection (footing)

	// Calculate deflection based on difference from optimal TWA
	angleDiff := absTWA - optimalTWA

	if angleDiff < 0 {
		// Pinching (sailing higher than optimal) - telltale lifts up
		deflectionFactor := math.Abs(angleDiff) / optimalTWA
		deflectionFactor = math.Min(deflectionFactor, 1.0)

		// Apply efficiency factor - less efficient = more deflection
		efficiencyFactor := 1.0 - math.Min(efficiency, 1.0)
		t.Angle = maxUpAngle * deflectionFactor * (0.5 + efficiencyFactor*0.5)
	} else if angleDiff > 0 {
		// Footing (sailing lower than optimal) - telltale drops down
		maxAngleRange := 180.0 - optimalTWA // Available range to foot off
		if maxAngleRange <= 0 {
			maxAngleRange = 90.0 // Fallback
		}

		deflectionFactor := angleDiff / maxAngleRange
		deflectionFactor = math.Min(deflectionFactor, 1.0)

		// Apply efficiency factor
		efficiencyFactor := 1.0 - math.Min(efficiency, 1.0)
		t.Angle = maxDownAngle * deflectionFactor * (0.5 + efficiencyFactor*0.5)
	}
}

// Draw renders the single red telltale on screen
func (t *Telltales) Draw(screen *ebiten.Image) {
	if !t.Visible {
		return
	}

	// Draw red filled circle at base (telltale sticker)
	const stickerRadius = 10.0
	vector.DrawFilledCircle(screen,
		float32(t.BaseX), float32(t.BaseY),
		stickerRadius,
		color.RGBA{255, 0, 0, 255}, false) // Red filled circle

	// Draw single red telltale
	endX := t.BaseX + t.Length*math.Cos(t.Angle*math.Pi/180)
	endY := t.BaseY + t.Length*math.Sin(t.Angle*math.Pi/180)

	vector.StrokeLine(screen,
		float32(t.BaseX), float32(t.BaseY),
		float32(endX), float32(endY),
		4.0, color.RGBA{255, 0, 0, 255}, false) // Red, 4px thick for visibility
}
