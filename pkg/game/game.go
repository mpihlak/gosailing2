package game

import (
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/mpihlak/ebiten-sailing/pkg/dashboard"
	"github.com/mpihlak/ebiten-sailing/pkg/game/objects"
	"github.com/mpihlak/ebiten-sailing/pkg/game/world"
	"github.com/mpihlak/ebiten-sailing/pkg/geometry"
	"github.com/mpihlak/ebiten-sailing/pkg/polars"
)

const (
	ScreenWidth  = 1280
	ScreenHeight = 720
	// Real world scale: 1 pixel = 1 meter for easier calculations
	PixelsPerMeter = 1.0
	WorldWidth     = 2000 // World is larger than screen
	WorldHeight    = 1500
	inputDelay     = 1 * time.Millisecond // Delay between keystroke readings
)

type GameState struct {
	Boat      *objects.Boat
	Arena     *world.Arena
	Wind      world.Wind
	Dashboard *dashboard.Dashboard
	CameraX   float64 // Camera offset for panning
	CameraY   float64
	lastInput time.Time // Last time input was processed
}

func NewGame() *GameState {
	wind := &world.ConstantWind{
		Direction: 0, // From North
		Speed:     10,
	}

	// Position starting line in center of world, optimized for 720p view
	// Starting line at Y = 600, with room above and below
	pinX := float64(WorldWidth/2 - 300)       // Pin end (left)
	committeeX := float64(WorldWidth/2 + 300) // Committee end (right)
	lineY := float64(600)

	// Boat starts 180 meters below pin end, sailing parallel to line towards committee boat
	boatStartX := pinX        // Aligned with pin end
	boatStartY := lineY + 180 // 180 meters below the line

	boat := &objects.Boat{
		Pos:     geometry.Point{X: boatStartX, Y: boatStartY},
		Heading: 90, // Sailing East (parallel to line, towards committee boat)
		Speed:   0,  // Will be set to target speed
		Polars:  &polars.RealisticPolar{},
		Wind:    wind,
	}

	// Initialize boat at full target speed for current heading and wind conditions
	windDir, windSpeed := wind.GetWind(boat.Pos)
	twa := boat.Heading - windDir
	if twa < -180 {
		twa += 360
	} else if twa > 180 {
		twa -= 360
	}
	targetSpeed := boat.Polars.GetBoatSpeed(twa, windSpeed)
	boat.Speed = targetSpeed

	// Set velocity components to match target speed in heading direction
	headingRad := boat.Heading * math.Pi / 180
	targetPixelSpeed := targetSpeed * 30.0 / 6.0 / 60.0 // speedScale / 60.0
	boat.VelX = targetPixelSpeed * math.Sin(headingRad)
	boat.VelY = -targetPixelSpeed * math.Cos(headingRad) // Y inverted
	arena := &world.Arena{
		Marks: []*world.Mark{
			{Pos: geometry.Point{X: pinX, Y: lineY}, Name: "Pin"},
			{Pos: geometry.Point{X: committeeX, Y: lineY}, Name: "Committee"},
		},
	}
	dash := &dashboard.Dashboard{
		Boat:      boat,
		Wind:      wind,
		StartTime: time.Now().Add(5 * time.Minute),
		LineStart: geometry.Point{X: pinX, Y: lineY},       // Pin end
		LineEnd:   geometry.Point{X: committeeX, Y: lineY}, // Committee end
	}

	// Initialize camera to show full starting area (center on starting line)
	cameraX := (pinX+committeeX)/2 - float64(ScreenWidth)/2 // Center line horizontally
	cameraY := lineY - float64(ScreenHeight)/2 + 100        // Show line and area below

	return &GameState{
		Boat:      boat,
		Arena:     arena,
		Wind:      wind,
		Dashboard: dash,
		CameraX:   cameraX,
		CameraY:   cameraY,
	}
}

func (g *GameState) Update() error {
	// Input handling with delay to prevent overturning
	if time.Since(g.lastInput) >= inputDelay {
		if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
			g.Boat.Heading -= 1
			g.lastInput = time.Now()
		}
		if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
			g.Boat.Heading += 1
			g.lastInput = time.Now()
		}
	}

	// Normalize heading
	if g.Boat.Heading < 0 {
		g.Boat.Heading += 360
	}
	if g.Boat.Heading >= 360 {
		g.Boat.Heading -= 360
	}

	g.Boat.Update()

	// Update camera to follow boat when it moves out of bounds
	g.updateCamera()

	return nil
}

// updateCamera pans the camera to keep the boat visible
func (g *GameState) updateCamera() {
	boatScreenX := g.Boat.Pos.X - g.CameraX
	boatScreenY := g.Boat.Pos.Y - g.CameraY

	// Camera margins - start panning when boat gets within this distance from edge
	margin := 100.0

	// Pan horizontally if boat is near screen edges
	if boatScreenX < margin {
		g.CameraX = g.Boat.Pos.X - margin
	} else if boatScreenX > float64(ScreenWidth)-margin {
		g.CameraX = g.Boat.Pos.X - (float64(ScreenWidth) - margin)
	}

	// Pan vertically if boat is near screen edges
	if boatScreenY < margin {
		g.CameraY = g.Boat.Pos.Y - margin
	} else if boatScreenY > float64(ScreenHeight)-margin {
		g.CameraY = g.Boat.Pos.Y - (float64(ScreenHeight) - margin)
	}

	// Clamp camera to world bounds
	g.CameraX = math.Max(0, math.Min(g.CameraX, float64(WorldWidth-ScreenWidth)))
	g.CameraY = math.Max(0, math.Min(g.CameraY, float64(WorldHeight-ScreenHeight)))
}

func (g *GameState) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 105, 148, 255}) // Blue for water

	// Apply camera transform
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-g.CameraX, -g.CameraY)

	// Create world image to draw everything on
	worldImage := ebiten.NewImage(WorldWidth, WorldHeight)
	worldImage.Fill(color.RGBA{0, 105, 148, 255}) // Blue for water

	// Draw arena (which includes marks) to world
	g.Arena.Draw(worldImage)

	// Draw boat (which includes its history trail) to world
	g.Boat.Draw(worldImage)

	// Draw the world to screen with camera offset
	screen.DrawImage(worldImage, op)

	// Draw dashboard directly to screen (UI always visible)
	g.Dashboard.Draw(screen)
}

func (g *GameState) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
