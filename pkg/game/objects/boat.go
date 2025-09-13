package objects

import (
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/mpihlak/ebiten-sailing/pkg/geometry"
)

const (
	maxHistoryPoints = 50
	historyInterval  = 200 * time.Millisecond
)

type Boat struct {
	Pos         geometry.Point // Center of the boat
	Heading     float64        // in degrees
	Speed       float64        // in knots
	History     []geometry.Point
	lastHistory time.Time
}

// GetBowPosition returns the position of the boat's bow (front tip)
func (b *Boat) GetBowPosition() geometry.Point {
	headingRad := b.Heading * math.Pi / 180
	bowDistance := 7.5 // Half the triangle height (15/2)

	return geometry.Point{
		X: b.Pos.X + bowDistance*math.Sin(headingRad),
		Y: b.Pos.Y - bowDistance*math.Cos(headingRad),
	}
}

func (b *Boat) Update() {
	// Convert heading to radians for math functions
	headingRad := b.Heading * math.Pi / 180

	// Move boat
	b.Pos.X += b.Speed * math.Sin(headingRad)
	b.Pos.Y -= b.Speed * math.Cos(headingRad) // Y is inverted in screen coordinates

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
	historyToShow := len(b.History) - 2
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
	height := 15.0 // 1.5x bigger than 10
	width := 7.5   // 1.5x bigger than 5

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
