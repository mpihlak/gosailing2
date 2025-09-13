package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/mpihlak/ebiten-sailing/pkg/game/objects"
	"github.com/mpihlak/ebiten-sailing/pkg/game/world"
	"github.com/mpihlak/ebiten-sailing/pkg/geometry"
)

const (
	ScreenWidth  = 800
	ScreenHeight = 600
)

type GameState struct {
	Boat  *objects.Boat
	Arena *world.Arena
	Wind  world.Wind
}

func NewGame() *GameState {
	boat := &objects.Boat{
		Pos:     geometry.Point{X: ScreenWidth / 2, Y: ScreenHeight / 2},
		Heading: 0,
		Speed:   2, // Initial speed
	}
	arena := &world.Arena{
		Marks: []*world.Mark{
			{Pos: geometry.Point{X: ScreenWidth / 2, Y: 50}, Name: "Upwind"},
			{Pos: geometry.Point{X: 100, Y: ScreenHeight - 100}, Name: "Pin"},
			{Pos: geometry.Point{X: ScreenWidth - 100, Y: ScreenHeight - 100}, Name: "Committee"},
		},
	}
	wind := &world.ConstantWind{
		Direction: 0, // From North
		Speed:     10,
	}
	return &GameState{
		Boat:  boat,
		Arena: arena,
		Wind:  wind,
	}
}

func (g *GameState) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.Boat.Heading -= 5
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.Boat.Heading += 5
	}

	g.Boat.Update()
	return nil
}

func (g *GameState) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 105, 148, 255}) // Blue for water

	// Draw boat
	ebitenutil.DrawCircle(screen, g.Boat.Pos.X, g.Boat.Pos.Y, 10, color.White)

	// Draw marks
	for _, mark := range g.Arena.Marks {
		ebitenutil.DrawRect(screen, mark.Pos.X-5, mark.Pos.Y-5, 10, 10, color.RGBA{255, 0, 0, 255})
		ebitenutil.DebugPrintAt(screen, mark.Name, int(mark.Pos.X)+15, int(mark.Pos.Y))
	}
}

func (g *GameState) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
