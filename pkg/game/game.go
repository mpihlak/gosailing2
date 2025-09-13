package game

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mpihlak/ebiten-sailing/pkg/dashboard"
	"github.com/mpihlak/ebiten-sailing/pkg/game/objects"
	"github.com/mpihlak/ebiten-sailing/pkg/game/world"
	"github.com/mpihlak/ebiten-sailing/pkg/geometry"
)

const (
	ScreenWidth  = 1280
	ScreenHeight = 720
)

type GameState struct {
	Boat      *objects.Boat
	Arena     *world.Arena
	Wind      world.Wind
	Dashboard *dashboard.Dashboard
}

func NewGame() *GameState {
	boat := &objects.Boat{
		Pos:     geometry.Point{X: ScreenWidth / 2, Y: ScreenHeight - 100},
		Heading: 90,
		Speed:   1, // Initial speed
	}
	arena := &world.Arena{
		Marks: []*world.Mark{
			{Pos: geometry.Point{X: ScreenWidth / 2, Y: 50}, Name: "Upwind"},
			{Pos: geometry.Point{X: 100, Y: ScreenHeight - 200}, Name: "Pin"},
			{Pos: geometry.Point{X: ScreenWidth - 100, Y: ScreenHeight - 200}, Name: "Committee"},
		},
	}
	wind := &world.ConstantWind{
		Direction: 0, // From North
		Speed:     10,
	}
	dash := &dashboard.Dashboard{
		Boat:      boat,
		Wind:      wind,
		StartTime: time.Now().Add(5 * time.Minute),
	}
	return &GameState{
		Boat:      boat,
		Arena:     arena,
		Wind:      wind,
		Dashboard: dash,
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

	// Draw arena (which includes marks)
	g.Arena.Draw(screen)

	// Draw boat (which includes its history trail)
	g.Boat.Draw(screen)

	// Draw dashboard
	g.Dashboard.Draw(screen)
}

func (g *GameState) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
