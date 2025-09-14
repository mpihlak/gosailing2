package dashboard

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/mpihlak/ebiten-sailing/pkg/game/objects"
	"github.com/mpihlak/ebiten-sailing/pkg/game/world"
	"github.com/mpihlak/ebiten-sailing/pkg/geometry"
)

type Dashboard struct {
	Boat      *objects.Boat
	Wind      world.Wind
	StartTime time.Time
	LineStart geometry.Point // Pin end of starting line
	LineEnd   geometry.Point // Committee end of starting line
}

// CalculateDistanceToLine calculates the perpendicular distance from boat's bow to the starting line
// Returns negative distance when boat is on the course side (above) of the line
func (d *Dashboard) CalculateDistanceToLine() float64 {
	bowPos := d.Boat.GetBowPosition()

	// Calculate line equation (Ax + By + C = 0)
	// For line from LineStart to LineEnd
	A := d.LineEnd.Y - d.LineStart.Y
	B := d.LineStart.X - d.LineEnd.X
	C := d.LineEnd.X*d.LineStart.Y - d.LineStart.X*d.LineEnd.Y

	// Calculate signed distance (without absolute value)
	signedDistance := (A*bowPos.X + B*bowPos.Y + C) / math.Sqrt(A*A+B*B)

	// For a horizontal line (A=0), the sign indicates which side:
	// Since the starting line is horizontal and course is above (lower Y),
	// we want negative when boat is above the line (on course side)
	// The current calculation gives positive when Y < lineY, so we need to negate
	return -signedDistance
}

// CalculateVMG calculates the current VMG (Velocity Made Good) towards wind
func (d *Dashboard) CalculateVMG() float64 {
	windDir, _ := d.Wind.GetWind(d.Boat.Pos)
	twa := d.Boat.Heading - windDir
	if twa < -180 {
		twa += 360
	} else if twa > 180 {
		twa -= 360
	}

	// VMG = Speed * cos(TWA)
	twaRad := twa * math.Pi / 180
	return d.Boat.Speed * math.Cos(twaRad)
}

// FindBestVMG finds the best VMG achievable for current sailing mode (beat or run)
func (d *Dashboard) FindBestVMG() float64 {
	windDir, windSpeed := d.Wind.GetWind(d.Boat.Pos)
	twa := d.Boat.Heading - windDir
	if twa < -180 {
		twa += 360
	} else if twa > 180 {
		twa -= 360
	}

	absTWA := math.Abs(twa)
	bestVMG := 0.0

	if absTWA < 90 {
		// Upwind sailing - find best beat VMG (positive VMG towards wind)
		for angle := 30.0; angle <= 90.0; angle += 1.0 {
			speed := d.Boat.Polars.GetBoatSpeed(angle, windSpeed)
			angleRad := angle * math.Pi / 180
			vmg := speed * math.Cos(angleRad)

			if vmg > bestVMG {
				bestVMG = vmg
			}
		}
	} else {
		// Downwind sailing - find best run VMG (negative VMG away from wind)
		for angle := 90.0; angle <= 180.0; angle += 1.0 {
			speed := d.Boat.Polars.GetBoatSpeed(angle, windSpeed)
			angleRad := angle * math.Pi / 180
			vmg := speed * math.Cos(angleRad)

			if vmg < bestVMG {
				bestVMG = vmg
			}
		}
	}

	return bestVMG
}

func (d *Dashboard) Draw(screen *ebiten.Image, raceStarted bool, isOCS bool, startTime time.Time) {
	windDir, windSpeed := d.Wind.GetWind(d.Boat.Pos)
	twa := d.Boat.Heading - windDir
	if twa < -180 {
		twa += 360
	} else if twa > 180 {
		twa -= 360
	}

	distanceToLine := d.CalculateDistanceToLine()
	currentVMG := d.CalculateVMG()
	targetVMG := d.FindBestVMG()

	msg := fmt.Sprintf(
		"Speed: %.1f kts\nHeading: %.0f°\nTWA: %.0f°\nTWS: %.1f kts\nDist to Line: %.0fm\nVMG: %.1f kts\nTarget VMG: %.1f kts",
		d.Boat.Speed, d.Boat.Heading, twa, windSpeed, distanceToLine, currentVMG, targetVMG,
	)

	ebitenutil.DebugPrintAt(screen, msg, screen.Bounds().Dx()-150, 10)

	// Race timer display
	if !raceStarted {
		remaining := time.Until(startTime)
		if remaining < 0 {
			remaining = 0
		}
		minutes := int(remaining.Minutes())
		seconds := int(remaining.Seconds()) % 60
		timerMsg := fmt.Sprintf("Start: %02d:%02d", minutes, seconds)
		ebitenutil.DebugPrintAt(screen, timerMsg, screen.Bounds().Dx()-150, 130)
	} else {
		ebitenutil.DebugPrintAt(screen, "RACE STARTED", screen.Bounds().Dx()-150, 130)
	}

	// OCS warning
	if isOCS && !raceStarted {
		// Draw red background rectangle for OCS warning
		ocsX := screen.Bounds().Dx() - 150
		ocsY := 150
		ocsWidth := 80
		ocsHeight := 15

		// Create red rectangle
		redRect := ebiten.NewImage(ocsWidth, ocsHeight)
		redRect.Fill(color.RGBA{255, 0, 0, 255}) // Red background

		// Draw red rectangle
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(ocsX), float64(ocsY))
		screen.DrawImage(redRect, op)

		// Draw white text on red background
		ebitenutil.DebugPrintAt(screen, "*** OCS ***", ocsX, ocsY)
	}
}
