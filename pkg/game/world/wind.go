package world

import "github.com/mpihlak/gosailing2/pkg/geometry"

type Wind interface {
	GetWind(pos geometry.Point) (direction, speed float64)
}

type ConstantWind struct {
	Direction float64
	Speed     float64
}

func (cw *ConstantWind) GetWind(_ geometry.Point) (float64, float64) {
	return cw.Direction, cw.Speed
}

// VariableWind provides wind that varies in strength across the course
type VariableWind struct {
	Direction  float64 // Wind direction (constant)
	LeftSpeed  float64 // Wind speed on left side (X=0)
	RightSpeed float64 // Wind speed on right side (X=WorldWidth)
	WorldWidth float64 // Width of the world for interpolation
}

func (vw *VariableWind) GetWind(pos geometry.Point) (float64, float64) {
	// Interpolate wind speed based on X position
	// X=0 (left) = LeftSpeed, X=WorldWidth (right) = RightSpeed
	xRatio := pos.X / vw.WorldWidth
	if xRatio < 0 {
		xRatio = 0
	} else if xRatio > 1 {
		xRatio = 1
	}

	// Linear interpolation between left and right speeds
	speed := vw.LeftSpeed + (vw.RightSpeed-vw.LeftSpeed)*xRatio

	return vw.Direction, speed
}
