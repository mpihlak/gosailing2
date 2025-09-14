package game

import (
	"image/color"
	"math"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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
	inputDelay     = 0 * time.Millisecond // Delay between keystroke readings
)

type GameState struct {
	Boat           *objects.Boat
	Arena          *world.Arena
	Wind           world.Wind
	Dashboard      *dashboard.Dashboard
	CameraX        float64 // Camera offset for panning
	CameraY        float64
	lastInput      time.Time // Last time input was processed
	isPaused       bool      // Game pause state
	lastPauseInput time.Time // Last time pause key was pressed
	// Race start timer (elapsed time based for pause support)
	timerDuration  time.Duration // Total duration for race start (1 minute)
	elapsedTime    time.Duration // Time elapsed since game start (only when not paused)
	lastUpdateTime time.Time     // Last time Update was called (for calculating delta)
	raceStarted    bool          // Whether the race has started
	// OCS detection
	isOCS bool // Whether boat is On Course Side
	// Line crossing tracking
	hasCrossedLine   bool          // Whether boat has crossed the starting line after race start
	lineCrossingTime time.Duration // When boat crossed the line (elapsed time)
	secondsLate      float64       // How many seconds late the boat was
	vmgAtCrossing    float64       // VMG when crossing the line
	vmgPercentage    float64       // VMG as percentage of optimal VMG
	// Restart banner
	showRestartBanner bool      // Whether to show restart banner
	restartBannerTime time.Time // When restart banner was triggered
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

	// Boat starts 180 meters below middle of line, sailing parallel to line towards committee boat
	boatStartX := (pinX + committeeX) / 2 // Middle of the starting line
	boatStartY := lineY + 180             // 180 meters below the line

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
		Boat:              boat,
		Arena:             arena,
		Wind:              wind,
		Dashboard:         dash,
		CameraX:           cameraX,
		CameraY:           cameraY,
		isPaused:          true,            // Start game in paused mode
		timerDuration:     1 * time.Minute, // Race starts after 1 minute
		elapsedTime:       0,               // No time elapsed yet
		lastUpdateTime:    time.Now(),      // Initialize update time
		raceStarted:       false,
		isOCS:             false,
		showRestartBanner: false,
		restartBannerTime: time.Time{},
	}
}

func (g *GameState) Update() error {
	// Handle quit key
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		os.Exit(0)
	}

	// Handle restart key
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		newGame := NewGame()
		*g = *newGame
		// Unpause and show restart banner
		g.isPaused = false
		g.showRestartBanner = true
		g.restartBannerTime = time.Now()
		return nil
	}

	// Handle timer jump key
	if inpututil.IsKeyJustPressed(ebiten.KeyJ) {
		// Jump timer forward by 10 seconds
		g.elapsedTime += 10 * time.Second
		// If this pushes us past race start, trigger race start
		if !g.raceStarted && g.elapsedTime >= g.timerDuration {
			g.raceStarted = true
		}
	}

	// Handle pause toggle
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.isPaused = !g.isPaused
		if !g.isPaused {
			// Reset last update time when unpausing to avoid large time jump
			g.lastUpdateTime = time.Now()
		}
	}

	// Don't update game logic when paused
	if g.isPaused {
		return nil
	}

	// Update elapsed time (only when not paused)
	now := time.Now()
	deltaTime := now.Sub(g.lastUpdateTime)
	g.elapsedTime += deltaTime
	g.lastUpdateTime = now

	// Hide restart banner after 2 seconds
	if g.showRestartBanner && time.Since(g.restartBannerTime) > 2*time.Second {
		g.showRestartBanner = false
	}

	// Check race start timer based on elapsed time
	if !g.raceStarted && g.elapsedTime >= g.timerDuration {
		g.raceStarted = true
	}

	// OCS detection - check if boat is above (course side of) the starting line
	// Starting line is at Y = 600, boat is OCS if Y < 600
	startLineY := 600.0
	if !g.raceStarted {
		// Before race start, check if boat crosses to course side
		if g.Boat.Pos.Y < startLineY {
			g.isOCS = true
		} else if g.isOCS && g.Boat.Pos.Y > startLineY {
			// Clear OCS only when boat dips below the line
			g.isOCS = false
		}
	} else {
		// After race start, only clear OCS if boat goes below the line
		// (boat can still be OCS from before race start)
		if g.isOCS && g.Boat.Pos.Y > startLineY {
			g.isOCS = false
		}

		// Line crossing detection after race start
		// Only count line crossing if boat is not currently OCS (has cleared OCS properly)
		if !g.hasCrossedLine && !g.isOCS {
			bowPos := g.Boat.GetBowPosition()
			// Check if bow crosses the starting line (from below to above)
			if bowPos.Y <= startLineY {
				g.hasCrossedLine = true
				g.lineCrossingTime = g.elapsedTime
				// Calculate how late the boat was (time after race start)
				g.secondsLate = (g.elapsedTime - g.timerDuration).Seconds()
				// Calculate VMG at crossing
				g.vmgAtCrossing = g.Dashboard.CalculateVMG()
				// Calculate VMG percentage (compare to optimal VMG)
				optimalVMG := g.Dashboard.FindBestVMG()
				if optimalVMG > 0 {
					g.vmgPercentage = (g.vmgAtCrossing / optimalVMG) * 100
				} else {
					g.vmgPercentage = 0
				}
			}
		}
	}

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
	g.Arena.Draw(worldImage, g.raceStarted)

	// Draw boat (which includes its history trail) to world
	g.Boat.Draw(worldImage)

	// Draw the world to screen with camera offset
	screen.DrawImage(worldImage, op)

	// Draw dashboard directly to screen (UI always visible)
	g.Dashboard.Draw(screen, g.raceStarted, g.isOCS, g.timerDuration, g.elapsedTime, g.hasCrossedLine, g.secondsLate, g.vmgPercentage)

	// Show START banner when race just started (for 3 seconds after race start)
	if g.raceStarted && g.elapsedTime-g.timerDuration < 3*time.Second {
		g.drawStartBanner(screen)
	}

	// Show RESTART banner when game was restarted
	if g.showRestartBanner {
		g.drawRestartBanner(screen)
	}

	// Draw help screen when paused
	if g.isPaused {
		g.drawHelpScreen(screen)
	}
}

// drawHelpScreen displays the help overlay when game is paused
func (g *GameState) drawHelpScreen(screen *ebiten.Image) {
	// Semi-transparent overlay
	overlay := ebiten.NewImage(ScreenWidth, ScreenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 180})
	screen.DrawImage(overlay, nil)

	// Help text
	helpText := `SAILING GAME - PAUSED

Controls:
  Left Arrow / A  - Turn Left
  Right Arrow / D - Turn Right
  Space           - Pause/Resume
  J               - Jump Timer +10 sec
  R               - Restart Game
  Q               - Quit Game

Dashboard:
  Speed     - Current boat speed
  Heading   - Boat direction (0-360°)
  TWA       - True Wind Angle
  TWS       - True Wind Speed
  VMG       - Velocity Made Good
  Target VMG - Best achievable VMG

Press SPACE to continue...`

	// Center the help text
	bounds := screen.Bounds()
	x := bounds.Dx()/2 - 200
	y := bounds.Dy()/2 - 150

	ebitenutil.DebugPrintAt(screen, helpText, x, y)
}

// drawStartBanner displays the START banner when race begins
func (g *GameState) drawStartBanner(screen *ebiten.Image) {
	bounds := screen.Bounds()

	// Semi-transparent overlay
	overlay := ebiten.NewImage(ScreenWidth, ScreenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 100})
	screen.DrawImage(overlay, nil)

	// START banner text
	startText := "*** RACE START! ***"

	// Center the text
	x := bounds.Dx()/2 - 80 // Approximate centering
	y := bounds.Dy()/2 - 20

	ebitenutil.DebugPrintAt(screen, startText, x, y)
}

// drawRestartBanner displays the RESTART banner when game is restarted
func (g *GameState) drawRestartBanner(screen *ebiten.Image) {
	bounds := screen.Bounds()

	// Semi-transparent overlay
	overlay := ebiten.NewImage(ScreenWidth, ScreenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 100})
	screen.DrawImage(overlay, nil)

	// RESTART banner text
	restartText := "*** RESTARTED ***"

	// Center the text
	x := bounds.Dx()/2 - 80 // Approximate centering
	y := bounds.Dy()/2 - 20

	ebitenutil.DebugPrintAt(screen, restartText, x, y)
}

func (g *GameState) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
