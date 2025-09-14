package dashboard

import (
	"fmt"
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

func (d *Dashboard) Draw(screen *ebiten.Image) {
	windDir, windSpeed := d.Wind.GetWind(d.Boat.Pos)
	twa := d.Boat.Heading - windDir
	if twa < -180 {
		twa += 360
	} else if twa > 180 {
		twa -= 360
	}

	distanceToLine := d.CalculateDistanceToLine()

	msg := fmt.Sprintf(
		"Speed: %.1f kts\nHeading: %.0f°\nTWA: %.0f°\nTWS: %.1f kts\nDist to Line: %.0fm",
		d.Boat.Speed, d.Boat.Heading, twa, windSpeed, distanceToLine,
	)

	ebitenutil.DebugPrintAt(screen, msg, screen.Bounds().Dx()-150, 10)

	// Countdown timer
	remaining := time.Until(d.StartTime)
	if remaining < 0 {
		remaining = 0
	}
	minutes := int(remaining.Minutes())
	seconds := int(remaining.Seconds()) % 60
	timerMsg := fmt.Sprintf("Start: %02d:%02d", minutes, seconds)
	ebitenutil.DebugPrintAt(screen, timerMsg, screen.Bounds().Dx()-150, 90)
}
