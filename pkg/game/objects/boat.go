package objects

import (
	"math"

	"github.com/mpihlak/ebiten-sailing/pkg/geometry"
)

type Boat struct {
	Pos     geometry.Point
	Heading float64 // in degrees
	Speed   float64 // in knots
}

func (b *Boat) Update() {
	// Convert heading to radians for math functions
	headingRad := b.Heading * math.Pi / 180

	// Move boat
	b.Pos.X += b.Speed * math.Sin(headingRad)
	b.Pos.Y -= b.Speed * math.Cos(headingRad) // Y is inverted in screen coordinates
}
