package game

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/mpihlak/gosailing2/pkg/dashboard"
	"github.com/mpihlak/gosailing2/pkg/game/objects"
	"github.com/mpihlak/gosailing2/pkg/game/world"
	"github.com/mpihlak/gosailing2/pkg/geometry"
	"github.com/mpihlak/gosailing2/pkg/polars"
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
	// Mobile controls
	mobileControls *MobileControls
	// Reusable images to avoid creating new ones every frame
	worldImage *ebiten.Image
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
	speedPercentage  float64       // Speed as percentage of target beat speed
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
		mobileControls:    NewMobileControls(ScreenWidth, ScreenHeight),
		worldImage:        ebiten.NewImage(WorldWidth, WorldHeight),
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
	// Process mobile touch input
	mobileInput := g.mobileControls.Update()

	// Handle quit key - different behavior for WASM vs standalone
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		if IsWASM() {
			// In WASM, pause the game and show help screen instead of quitting
			g.isPaused = true
			return nil
		} else {
			// In standalone, return error to exit the application
			return fmt.Errorf("game quit by user")
		}
	}

	// Handle restart key (keyboard or mobile)
	if inpututil.IsKeyJustPressed(ebiten.KeyR) || mobileInput.RestartPressed {
		newGame := NewGame()
		*g = *newGame
		// Unpause and show restart banner
		g.isPaused = false
		g.showRestartBanner = true
		g.restartBannerTime = time.Now()
		return nil
	}

	// Handle timer jump key (keyboard or mobile)
	if inpututil.IsKeyJustPressed(ebiten.KeyJ) || mobileInput.TimerJumpPressed {
		// Jump timer forward by 10 seconds
		g.elapsedTime += 10 * time.Second
		// If this pushes us past race start, trigger race start
		if !g.raceStarted && g.elapsedTime >= g.timerDuration {
			g.raceStarted = true
		}
	}

	// Handle pause toggle (keyboard or mobile)
	pauseTogglePressed := inpututil.IsKeyJustPressed(ebiten.KeySpace) || mobileInput.PausePressed

	// On mobile, any touch when paused should unpause (except on buttons)
	if g.isPaused && g.mobileControls.hasTouchInput {
		justPressedTouchIDs := inpututil.AppendJustPressedTouchIDs(nil)
		for _, touchID := range justPressedTouchIDs {
			x, y := ebiten.TouchPosition(touchID)
			// Only unpause if touch is not on any mobile control buttons
			if !g.mobileControls.pauseButton.Contains(x, y) &&
				!g.mobileControls.menuButton.Contains(x, y) &&
				!g.mobileControls.restartButton.Contains(x, y) &&
				!g.mobileControls.timerButton.Contains(x, y) {
				pauseTogglePressed = true
				break
			}
		}
	}

	if pauseTogglePressed {
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
				// Calculate speed at crossing as percentage of target beat speed
				_, windSpeed := g.Wind.GetWind(g.Boat.Pos)
				// Target beat speed is typically at 45-50 degree TWA - use 45 degrees
				targetBeatSpeed := g.Boat.Polars.GetBoatSpeed(45.0, windSpeed)
				if targetBeatSpeed > 0 {
					g.speedPercentage = (g.Boat.Speed / targetBeatSpeed) * 100
				} else {
					g.speedPercentage = 0
				}
			}
		}
	}

	// Input handling with delay to prevent overturning
	if time.Since(g.lastInput) >= inputDelay {
		// Check keyboard input
		keyboardLeft := ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA)
		keyboardRight := ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD)

		// Combine keyboard and mobile input
		if keyboardLeft || mobileInput.TurnLeft {
			g.Boat.Heading -= 1
			g.lastInput = time.Now()
		}
		if keyboardRight || mobileInput.TurnRight {
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

	// Clear and redraw world image (reuse existing image instead of creating new one)
	g.worldImage.Fill(color.RGBA{0, 105, 148, 255}) // Blue for water

	// Draw arena (which includes marks) to world
	g.Arena.Draw(g.worldImage, g.raceStarted)

	// Draw boat (which includes its history trail) to world
	g.Boat.Draw(g.worldImage)

	// Draw the world to screen with camera offset
	screen.DrawImage(g.worldImage, op)

	// Draw dashboard directly to screen (UI always visible)
	g.Dashboard.Draw(screen, g.raceStarted, g.isOCS, g.timerDuration, g.elapsedTime, g.hasCrossedLine, g.secondsLate, g.speedPercentage)

	// Draw mobile controls (only visible on touch devices)
	g.mobileControls.Draw(screen)

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
	// Draw semi-transparent overlay using vector instead of creating new image
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 180}, false)

	var helpText string

	// Check if we're on mobile (touch input detected)
	if g.mobileControls.hasTouchInput {
		// Mobile help text - focus on game explanation, not controls
		helpText = `SAILING GAME - PAUSED

How to Play:
  🎯 Start racing when the timer reaches zero
  ⚠️  Avoid being OCS (On Course Side) at start
  🏁 Cross the starting line after race begins
  💨 Use wind angles for optimal speed

Touch Controls:
  Left/Right sides  - Steer the boat
  Pause button (⏸)  - Pause/Resume game
  Menu button (☰)   - Show restart & timer options

Dashboard Info:
  Speed     - Current boat speed in knots
  Heading   - Boat direction (0-360°)
  TWA       - True Wind Angle (-180 to +180°)
  TWS       - True Wind Speed in knots
  VMG       - Velocity Made Good (speed toward wind)
  Target VMG - Best achievable VMG for conditions
  Dist to Line - Distance to starting line

Tap anywhere to continue...`
	} else {
		// Desktop help text - include keyboard shortcuts
		quitText := "Quit Game"
		if IsWASM() {
			quitText = "Pause Game"
		}

		helpText = fmt.Sprintf(`SAILING GAME - PAUSED

How to Play:
  🎯 Start racing when the timer reaches zero
  ⚠️  Avoid being OCS (On Course Side) at start
  🏁 Cross the starting line after race begins
  💨 Use wind angles for optimal speed

Controls:
  Left Arrow / A  - Turn Left
  Right Arrow / D - Turn Right
  Space           - Pause/Resume
  J               - Jump Timer +10 sec
  R               - Restart Game
  Q               - %s

Dashboard Info:
  Speed     - Current boat speed in knots
  Heading   - Boat direction (0-360°)
  TWA       - True Wind Angle (-180 to +180°)
  TWS       - True Wind Speed in knots
  VMG       - Velocity Made Good (speed toward wind)
  Target VMG - Best achievable VMG for conditions
  Dist to Line - Distance to starting line

Press SPACE to continue...`, quitText)
	}

	// Center the help text
	bounds := screen.Bounds()
	x := bounds.Dx()/2 - 200
	y := bounds.Dy()/2 - 150

	ebitenutil.DebugPrintAt(screen, helpText, x, y)
}

// drawStartBanner displays the START banner when race begins
func (g *GameState) drawStartBanner(screen *ebiten.Image) {
	bounds := screen.Bounds()

	// Semi-transparent overlay using vector drawing
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 100}, false)

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

	// Semi-transparent overlay using vector drawing
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 100}, false)

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
