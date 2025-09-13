package world

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/mpihlak/ebiten-sailing/pkg/geometry"
)

type Mark struct {
	Pos  geometry.Point
	Name string
}

func (m *Mark) Draw(screen *ebiten.Image) {
	ebitenutil.DrawRect(screen, m.Pos.X-5, m.Pos.Y-5, 10, 10, color.RGBA{255, 0, 0, 255})
	ebitenutil.DebugPrintAt(screen, m.Name, int(m.Pos.X)+15, int(m.Pos.Y))
}

type Arena struct {
	Marks []*Mark
}

func (a *Arena) Draw(screen *ebiten.Image) {
	for _, mark := range a.Marks {
		mark.Draw(screen)
	}
}
