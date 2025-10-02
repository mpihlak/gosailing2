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
	// Wobble animation
	elapsedTime float64 // Time elapsed for wobble animation
	wobblePhase float64 // Phase offset for wobble (randomized)
}

// NewTelltales creates a new telltales instance
func NewTelltales(screenWidth, screenHeight int) *Telltales {
	return &Telltales{
		Length:      75.0,
		BaseX:       float64(screenWidth/2 - 50), // Left of center
		BaseY:       80.0,                        // Below timer and OCS warning
		Angle:       0.0,                         // Start horizontal
		Visible:     true,                        // Always visible now
		elapsedTime: 0.0,
		wobblePhase: math.Pi * 0.3, // Slight phase offset for natural look
	}
}

// Update calculates telltale position based on boat performance
func (t *Telltales) Update(boat *objects.Boat, wind world.Wind, dashboard *dashboard.Dashboard) {
	// Update animation time (assuming 60 FPS, so ~16.67ms per frame)
	t.elapsedTime += 1.0 / 60.0

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

	// Calculate base telltale angle based on VMG efficiency
	t.calculateTelltaleAngle(absTWA, optimalTWA, efficiency, windSpeed)
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

// calculateTelltaleAngle determines telltale angle based on VMG efficiency with natural wobble
func (t *Telltales) calculateTelltaleAngle(absTWA, optimalTWA, efficiency, windSpeed float64) {
	// Calculate base angle from VMG efficiency using aggressive response curve
	baseAngle := 0.0

	// Clamp efficiency to reasonable range
	efficiency = math.Max(0.0, math.Min(efficiency, 1.2)) // Allow slight over-efficiency

	// Aggressive telltale deflection based on VMG efficiency
	var deflectionAngle float64

	if efficiency >= 0.95 {
		// Very efficient sailing (95%+) - telltale nearly horizontal
		deflectionAngle = 0.0
	} else if efficiency >= 0.75 {
		// Good sailing (75-95%) - linear interpolation from 0° to 45°
		factor := (0.95 - efficiency) / (0.95 - 0.75) // 0 to 1 as efficiency drops from 95% to 75%
		deflectionAngle = 45.0 * factor
	} else if efficiency >= 0.50 {
		// Poor sailing (50-75%) - steeper curve from 45° to 75°
		factor := (0.75 - efficiency) / (0.75 - 0.50) // 0 to 1 as efficiency drops from 75% to 50%
		deflectionAngle = 45.0 + 30.0*factor          // 45° to 75°
	} else {
		// Very poor sailing (below 50%) - nearly vertical at 85°
		factor := math.Max(0.0, efficiency) / 0.50 // 0 to 1 as efficiency goes from 0% to 50%
		deflectionAngle = 85.0 - 10.0*factor       // 85° down to 75°
	}

	// Determine direction based on sailing mode relative to optimal TWA
	angleDiff := absTWA - optimalTWA

	if math.Abs(angleDiff) < 2.0 {
		// Very close to optimal - minimal deflection regardless of efficiency
		sign := 1.0
		if angleDiff < 0 {
			sign = -1.0
		}
		baseAngle = deflectionAngle * 0.2 * sign
	} else if angleDiff < 0 {
		// Pinching (sailing higher than optimal) - telltale lifts up (negative angle)
		baseAngle = -deflectionAngle
	} else {
		// Footing (sailing lower than optimal) - telltale drops down (positive angle)
		baseAngle = deflectionAngle * 0.7 // Slightly less dramatic for footing
	}

	// Add natural wobble animation
	// Wobble frequency and amplitude based on wind speed and efficiency
	wobbleFrequency := 2.0 + windSpeed*0.1                       // Higher wind = faster wobble
	wobbleAmplitude := 3.0 + (1.0-math.Min(efficiency, 1.0))*2.0 // Less efficient = more wobble

	// Create complex wobble with multiple sine waves for natural movement
	wobbleAngle1 := math.Sin(t.elapsedTime*wobbleFrequency+t.wobblePhase) * wobbleAmplitude
	wobbleAngle2 := math.Sin(t.elapsedTime*wobbleFrequency*1.7+t.wobblePhase*1.3) * wobbleAmplitude * 0.3
	wobbleAngle3 := math.Sin(t.elapsedTime*wobbleFrequency*0.6+t.wobblePhase*0.7) * wobbleAmplitude * 0.5

	totalWobble := wobbleAngle1 + wobbleAngle2 + wobbleAngle3

	// Combine base angle with wobble
	t.Angle = baseAngle + totalWobble
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
