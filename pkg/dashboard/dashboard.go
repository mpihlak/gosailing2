package dashboard

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/mpihlak/ebiten-sailing/pkg/game/objects"
	"github.com/mpihlak/ebiten-sailing/pkg/game/world"
)

type Dashboard struct {
	Boat      *objects.Boat
	Wind      world.Wind
	StartTime time.Time
}

func (d *Dashboard) Draw(screen *ebiten.Image) {
	windDir, windSpeed := d.Wind.GetWind(d.Boat.Pos)
	twa := d.Boat.Heading - windDir
	if twa < -180 {
		twa += 360
	} else if twa > 180 {
		twa -= 360
	}

	msg := fmt.Sprintf(
		"Speed: %.1f kts\nHeading: %.0f°\nTWA: %.0f°\nTWS: %.1f kts",
		d.Boat.Speed, d.Boat.Heading, twa, windSpeed,
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
	ebitenutil.DebugPrintAt(screen, timerMsg, screen.Bounds().Dx()-150, 70)
}
