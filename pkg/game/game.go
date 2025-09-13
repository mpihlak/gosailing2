package game

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
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

	// Draw boat
	ebitenutil.DrawCircle(screen, g.Boat.Pos.X, g.Boat.Pos.Y, 10, color.White)

	// Draw boat history
	for _, p := range g.Boat.History {
		ebitenutil.DrawCircle(screen, p.X, p.Y, 2, color.RGBA{173, 216, 230, 150})
	}

	// Draw marks
	for _, mark := range g.Arena.Marks {
		ebitenutil.DrawRect(screen, mark.Pos.X-5, mark.Pos.Y-5, 10, 10, color.RGBA{255, 0, 0, 255})
		ebitenutil.DebugPrintAt(screen, mark.Name, int(mark.Pos.X)+15, int(mark.Pos.Y))
	}

	g.Dashboard.Draw(screen)
}

func (g *GameState) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
