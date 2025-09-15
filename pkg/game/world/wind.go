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
