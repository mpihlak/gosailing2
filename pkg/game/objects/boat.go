package objects

import (
	"math"
	"time"

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
