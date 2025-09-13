package world

import "github.com/mpihlak/ebiten-sailing/pkg/geometry"

type Mark struct {
	Pos  geometry.Point
	Name string
}

type Arena struct {
	Marks []*Mark
}
