package objects

import (
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/mpihlak/ebiten-sailing/pkg/game/world"
	"github.com/mpihlak/ebiten-sailing/pkg/geometry"
	"github.com/mpihlak/ebiten-sailing/pkg/polars"
)

const (
	maxHistoryPoints = 50
	historyInterval  = 200 * time.Millisecond
	boatHeight       = 15.0       // Triangle height
	boatWidth        = 7.5        // Triangle width
	speedScale       = 30.0 / 6.0 // Pixels per second per knot (10 pixels/sec at 6 knots)
)

type Boat struct {
	Pos         geometry.Point // Center of the boat
	Heading     float64        // in degrees
	Speed       float64        // in knots
	History     []geometry.Point
	lastHistory time.Time
	Polars      polars.Polars // Polar performance data
	Wind        world.Wind    // Wind interface to get wind conditions
}

// GetBowPosition returns the position of the boat's bow (front tip)
func (b *Boat) GetBowPosition() geometry.Point {
	headingRad := b.Heading * math.Pi / 180
	bowDistance := boatHeight / 2

	return geometry.Point{
		X: b.Pos.X + bowDistance*math.Sin(headingRad),
		Y: b.Pos.Y - bowDistance*math.Cos(headingRad),
	}
}

func (b *Boat) Update() {
	// Get wind conditions at boat position
	windDir, windSpeed := b.Wind.GetWind(b.Pos)

	// Calculate True Wind Angle (TWA)
	twa := b.Heading - windDir
	if twa < -180 {
		twa += 360
	} else if twa > 180 {
		twa -= 360
	}

	// Update boat speed based on polars
	b.Speed = b.Polars.GetBoatSpeed(twa, windSpeed)

	// Convert heading to radians for math functions
	headingRad := b.Heading * math.Pi / 180

	// Scale speed from knots to pixels per frame (assuming 60 FPS)
	pixelSpeed := b.Speed * speedScale / 60.0

	// Move boat
	b.Pos.X += pixelSpeed * math.Sin(headingRad)
	b.Pos.Y -= pixelSpeed * math.Cos(headingRad) // Y is inverted in screen coordinates

	// Add to history
	if time.Since(b.lastHistory) >= historyInterval {
		b.History = append(b.History, b.Pos)
		b.lastHistory = time.Now()

		// Cap history at maxHistoryPoints
		if len(b.History) > maxHistoryPoints {
			b.History = b.History[1:]
		}
	}
}

func (b *Boat) Draw(screen *ebiten.Image) {
	// Draw boat history (skip the last 2 points to avoid overlap with boat)
	historyToShow := len(b.History) - 1
	if historyToShow < 0 {
		historyToShow = 0
	}

	for i := 0; i < historyToShow; i++ {
		p := b.History[i]
		ebitenutil.DrawCircle(screen, p.X, p.Y, 2, color.RGBA{173, 216, 230, 150})
	}

	// Draw boat as triangle pointing towards heading
	headingRad := b.Heading * math.Pi / 180

	// Triangle dimensions
	height := boatHeight
	width := boatWidth

	// Calculate triangle vertices relative to boat center position
	// Bow (tip) is forward from center, stern (base) is behind center
	bowDistance := height / 2
	sternDistance := height / 2

	// Bow position (front tip)
	bowX := b.Pos.X + bowDistance*math.Sin(headingRad)
	bowY := b.Pos.Y - bowDistance*math.Cos(headingRad)

	// Stern center position (back center)
	sternX := b.Pos.X - sternDistance*math.Sin(headingRad)
	sternY := b.Pos.Y + sternDistance*math.Cos(headingRad)

	// Left and right stern points
	leftX := sternX - (width/2)*math.Cos(headingRad)
	leftY := sternY - (width/2)*math.Sin(headingRad)

	rightX := sternX + (width/2)*math.Cos(headingRad)
	rightY := sternY + (width/2)*math.Sin(headingRad)

	// Draw triangle using lines
	ebitenutil.DrawLine(screen, bowX, bowY, leftX, leftY, color.White)
	ebitenutil.DrawLine(screen, leftX, leftY, rightX, rightY, color.White)
	ebitenutil.DrawLine(screen, rightX, rightY, bowX, bowY, color.White)
}
