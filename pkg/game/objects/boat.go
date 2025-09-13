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
	Pos         geometry.Point
	Heading     float64 // in degrees
	Speed       float64 // in knots
	History     []geometry.Point
	lastHistory time.Time
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
	// Draw boat history
	for _, p := range b.History {
		ebitenutil.DrawCircle(screen, p.X, p.Y, 2, color.RGBA{173, 216, 230, 150})
	}

	// Draw boat as triangle pointing towards heading
	headingRad := b.Heading * math.Pi / 180

	// Triangle dimensions
	height := 10.0
	width := 5.0

	// Calculate triangle vertices relative to boat position
	// Tip is at boat position, base is behind
	tipX := b.Pos.X
	tipY := b.Pos.Y

	// Base vertices (behind the tip)
	baseX := b.Pos.X - height*math.Sin(headingRad)
	baseY := b.Pos.Y + height*math.Cos(headingRad)

	// Left and right base points
	leftX := baseX - (width/2)*math.Cos(headingRad)
	leftY := baseY - (width/2)*math.Sin(headingRad)

	rightX := baseX + (width/2)*math.Cos(headingRad)
	rightY := baseY + (width/2)*math.Sin(headingRad)

	// Draw triangle using lines
	ebitenutil.DrawLine(screen, tipX, tipY, leftX, leftY, color.White)
	ebitenutil.DrawLine(screen, leftX, leftY, rightX, rightY, color.White)
	ebitenutil.DrawLine(screen, rightX, rightY, tipX, tipY, color.White)
}
