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
	WorldWidth     = 2000                 // World is larger than screen
	WorldHeight    = 3000                 // Expanded to accommodate upwind mark at Y=-1200
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
	timerDuration  time.Duration // Total duration for race start (30 seconds)
	elapsedTime    time.Duration // Time elapsed since game start (only when not paused)
	lastUpdateTime time.Time     // Last time Update was called (for calculating delta)
	raceStarted    bool          // Whether the race has started
	raceTimer      time.Duration // Time since race started (counts up from 0)
	// OCS detection
	isOCS bool // Whether boat is On Course Side
	// Line crossing tracking
	hasCrossedLine   bool           // Whether boat has crossed the starting line after race start
	lineCrossingTime time.Duration  // When boat crossed the line (elapsed time)
	secondsLate      float64        // How many seconds late the boat was
	vmgAtCrossing    float64        // VMG when crossing the line
	speedPercentage  float64        // Speed as percentage of target beat speed
	prevBowPos       geometry.Point // Previous frame's bow position for crossing detection
	// Mark rounding tracking
	markRoundingPhase1 bool // Sailed past mark (south to north)
	markRoundingPhase2 bool // Travelled to left (east to west while north)
	markRoundingPhase3 bool // Sailed below mark (north to south)
	markRounded        bool // All three phases completed
	// Race completion
	raceFinished     bool          // Whether boat has finished the race
	finishTime       time.Duration // Race time when boat finished
	showFinishBanner bool          // Whether to show finish banner
	finishBannerTime time.Time     // When finish banner was triggered
	// Restart banner
	showRestartBanner bool      // Whether to show restart banner
	restartBannerTime time.Time // When restart banner was triggered
}

func NewGame() *GameState {
	wind := world.NewOscillatingWind(
		14,         // 14 kts on left side
		8,          // 8 kts on right side
		WorldWidth, // Use world width for interpolation
	)

	// Position starting line in center of world, optimized for 720p view
	// Starting line at Y = 2400, shorter line (400m instead of 600m)
	// Upwind mark positioned to be immediately visible at top of screen
	pinX := float64(WorldWidth/2 - 200)       // Pin end (left) - shorter line
	committeeX := float64(WorldWidth/2 + 200) // Committee end (right) - shorter line
	lineY := float64(2400)                    // Positioned to accommodate upwind mark

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

	// Calculate upwind mark position (positioned to be visible at top of screen)
	upwindMarkX := (pinX + committeeX) / 2             // Center of starting line
	upwindMarkY := lineY - float64(ScreenHeight) + 100 // Visible at top of screen with margin

	arena := &world.Arena{
		Marks: []*world.Mark{
			{Pos: geometry.Point{X: pinX, Y: lineY}, Name: "Pin"},
			{Pos: geometry.Point{X: committeeX, Y: lineY}, Name: "Committee"},
			{Pos: geometry.Point{X: upwindMarkX, Y: upwindMarkY}, Name: "Upwind"},
		},
	}
	dash := &dashboard.Dashboard{
		Boat:       boat,
		Wind:       wind,
		StartTime:  time.Now().Add(5 * time.Minute),
		LineStart:  geometry.Point{X: pinX, Y: lineY},              // Pin end
		LineEnd:    geometry.Point{X: committeeX, Y: lineY},        // Committee end
		UpwindMark: geometry.Point{X: upwindMarkX, Y: upwindMarkY}, // Upwind mark
	}

	// Initialize camera to show full starting area (center on starting line)
	cameraX := (pinX+committeeX)/2 - float64(ScreenWidth)/2 // Center line horizontally
	cameraY := lineY - float64(ScreenHeight)/2 + 50         // Show line and upwind mark

	return &GameState{
		Boat:           boat,
		Arena:          arena,
		Wind:           wind,
		Dashboard:      dash,
		CameraX:        cameraX,
		CameraY:        cameraY,
		mobileControls: NewMobileControls(ScreenWidth, ScreenHeight),
		worldImage:     ebiten.NewImage(WorldWidth, WorldHeight),
		isPaused:       true,             // Start game in paused mode
		timerDuration:  30 * time.Second, // Race starts after 30 seconds
		elapsedTime:    0,                // No time elapsed yet
		lastUpdateTime: time.Now(),       // Initialize update time
		raceStarted:    false,
		raceTimer:      0, // Race timer starts at 0
		isOCS:          false,
		prevBowPos:     geometry.Point{X: boatStartX, Y: boatStartY}, // Initialize to boat start position
		// Mark rounding state
		markRoundingPhase1: false,
		markRoundingPhase2: false,
		markRoundingPhase3: false,
		markRounded:        false,
		// Race completion state
		raceFinished:      false,
		finishTime:        0,
		showFinishBanner:  false,
		finishBannerTime:  time.Time{},
		showRestartBanner: false,
		restartBannerTime: time.Time{},
	}
}

func (g *GameState) Update() error {
	// Process mobile touch input
	g.mobileControls.Update()
	mobileInput := g.mobileControls.GetMobileInput()

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

	// Handle 'c' key to toggle mobile controls display for testing
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		g.mobileControls.ToggleControlsOverride()
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

	// Handle 'J' key to jump timer forward by 10 seconds (only before race starts)
	if inpututil.IsKeyJustPressed(ebiten.KeyJ) && !g.raceStarted {
		g.elapsedTime += 10 * time.Second
		// Make sure we don't go past the timer duration
		if g.elapsedTime > g.timerDuration {
			g.elapsedTime = g.timerDuration
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
				!g.mobileControls.leftButton.Contains(x, y) &&
				!g.mobileControls.rightButton.Contains(x, y) &&
				!g.mobileControls.restartButton.Contains(x, y) {
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

	// Update wind oscillations (only when not paused)
	if oscillatingWind, ok := g.Wind.(*world.OscillatingWind); ok {
		oscillatingWind.Update()
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

	// Hide finish banner after 5 seconds
	if g.showFinishBanner && time.Since(g.finishBannerTime) > 5*time.Second {
		g.showFinishBanner = false
	}

	// Check race start timer based on elapsed time
	if !g.raceStarted && g.elapsedTime >= g.timerDuration {
		g.raceStarted = true
		g.raceTimer = 0 // Initialize race timer when race starts
	}

	// Update race timer if race has started but not finished
	if g.raceStarted && !g.raceFinished {
		g.raceTimer += deltaTime
	}

	// OCS detection and clearing - check if boat's bow is above (course side of) the starting line
	// Starting line is at Y = 2400, boat is OCS if bow crosses between pin and committee boat before race start
	startLineY := 2400.0
	bowPos := g.Boat.GetBowPosition()

	if !g.raceStarted {
		// Before race start, boat goes OCS if bow crosses the line between pin and committee boat
		if bowPos.Y <= startLineY && g.isWithinLineBounds(bowPos) {
			g.isOCS = true
		}
		// Clear OCS only when boat crosses back below the line between pin and committee boat
		if g.isOCS && bowPos.Y > startLineY && g.isWithinLineBounds(bowPos) {
			g.isOCS = false
		}
	} else {
		// After race start, OCS can still be cleared by crossing back below the line between pin and committee boat
		if g.isOCS && bowPos.Y > startLineY && g.isWithinLineBounds(bowPos) {
			g.isOCS = false
		}

		// Line crossing detection after race start
		// Only count line crossing if boat is not currently OCS (has cleared OCS properly)
		if !g.hasCrossedLine && !g.isOCS {
			// Check if bow crosses the Y coordinate from below (prevBowPos.Y > startLineY) to above (bowPos.Y <= startLineY)
			// AND the boat is within line bounds at the moment of crossing
			if g.prevBowPos.Y > startLineY && bowPos.Y <= startLineY && g.isWithinLineBounds(bowPos) {
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

		// Mark rounding detection (only if race has started and boat has crossed starting line)
		if g.hasCrossedLine && !g.raceFinished {
			g.updateMarkRounding()
		}

		// Finish line detection (only if boat has started and rounded the mark)
		if g.hasCrossedLine && g.markRounded && !g.raceFinished {
			g.checkFinishLineCrossing()
		}
	}

	// Update previous bow position for next frame's crossing detection
	g.prevBowPos = bowPos

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
	margin := 200.0

	// Pan horizontally if boat is near screen edges
	if boatScreenX < margin {
		g.CameraX = g.Boat.Pos.X - margin
	} else if boatScreenX > float64(ScreenWidth)-margin {
		g.CameraX = g.Boat.Pos.X - (float64(ScreenWidth) - margin)
	}

	// Pan vertically if boat is near screen edges (200px from top/bottom)
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
	g.Arena.Draw(g.worldImage, g.raceStarted, g.Wind)

	// Draw boat (which includes its history trail) to world
	g.Boat.Draw(g.worldImage)

	// Draw the world to screen with camera offset
	screen.DrawImage(g.worldImage, op)

	// Draw dashboard directly to screen (UI always visible)
	g.Dashboard.Draw(screen, g.raceStarted, g.isOCS, g.timerDuration, g.elapsedTime, g.hasCrossedLine, g.secondsLate, g.speedPercentage, g.markRounded, g.raceFinished)

	// Draw race timer at top center (when race hasn't started)
	g.drawRaceTimer(screen)

	// Draw OCS warning below timer
	g.drawOCSWarning(screen)

	// Draw mobile controls (only visible on touch devices)
	g.mobileControls.Draw(screen, g.isPaused)

	// Show START banner when race just started (for 3 seconds after race start)
	if g.raceStarted && g.elapsedTime-g.timerDuration < 3*time.Second {
		g.drawStartBanner(screen)
	}

	// Show RESTART banner when game was restarted
	if g.showRestartBanner {
		g.drawRestartBanner(screen)
	}

	// Show FINISH banner when race is finished
	if g.showFinishBanner {
		g.drawFinishBanner(screen)
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
		helpText = `Go Sailing! - PAUSED

How to Play:
* Start racing when the timer reaches zero
* Avoid being OCS (On Course Side) at start
* Cross the starting line after race begins
* Sail upwind and round the mark (leave to port)
* Return and cross finish line to complete race
* Use wind angles for optimal speed

Use Touch Controls to turn left/right, pause or restart.

Tap anywhere to continue...`
	} else {
		// Desktop help text - include keyboard shortcuts
		quitText := "Quit Game"
		if IsWASM() {
			quitText = "Pause Game"
		}

		helpText = fmt.Sprintf(`SAILING GAME - PAUSED

How to Play:
* Start racing when the timer reaches zero
* Avoid being OCS (On Course Side) at start
* Cross the starting line after race begins
* Sail upwind and round the mark (leave to port)
* Return and cross finish line to complete race
* Use wind angles for optimal speed

Controls:
  Left Arrow / A  - Turn Left
  Right Arrow / D - Turn Right
  Space           - Pause/Resume
  J               - Jump Timer +10 sec (pre start)
  R               - Restart Game
  C               - Toggle Touch Controls (testing)
  Q               - %s

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

// drawRaceTimer displays the race countdown timer or race time at the top center of the screen
func (g *GameState) drawRaceTimer(screen *ebiten.Image) {
	bounds := screen.Bounds()
	y := 20 // Top of screen with some margin

	if !g.raceStarted {
		// Show countdown timer before race starts
		remaining := g.timerDuration - g.elapsedTime
		if remaining < 0 {
			remaining = 0
		}
		minutes := int(remaining.Minutes())
		// Use ceiling for seconds to avoid showing 0 when there's still time left
		seconds := int(math.Ceil(remaining.Seconds())) % 60
		// Special case: if we're showing 0 minutes and 0 seconds but there's still time, show 1 second
		if minutes == 0 && seconds == 0 && remaining > 0 {
			seconds = 1
		}

		// Create larger, more visible timer display
		timerText := fmt.Sprintf("%02d:%02d", minutes, seconds)

		// Position at top center of screen
		x := bounds.Dx()/2 - 30 // Center horizontally (approximate for timer text)

		// Draw timer text
		ebitenutil.DebugPrintAt(screen, timerText, x, y)

		// Add "START IN:" label above the timer
		labelText := "START IN:"
		labelX := bounds.Dx()/2 - 35 // Center the label
		labelY := y - 15             // Above the timer
		ebitenutil.DebugPrintAt(screen, labelText, labelX, labelY)
	} else {
		// Show race timer after race starts
		minutes := int(g.raceTimer.Minutes())
		seconds := int(g.raceTimer.Seconds()) % 60

		// Create race timer display
		timerText := fmt.Sprintf("%02d:%02d", minutes, seconds)

		// Position at top center of screen
		x := bounds.Dx()/2 - 30 // Center horizontally (approximate for timer text)

		// Draw timer text
		ebitenutil.DebugPrintAt(screen, timerText, x, y)

		// Add "RACE TIME:" label above the timer
		labelText := "RACE TIME:"
		labelX := bounds.Dx()/2 - 42 // Center the label (slightly wider than "START IN:")
		labelY := y - 15             // Above the timer
		ebitenutil.DebugPrintAt(screen, labelText, labelX, labelY)
	}
}

// drawOCSWarning displays the OCS warning below the race timer
func (g *GameState) drawOCSWarning(screen *ebiten.Image) {
	// Only show OCS warning when boat is OCS
	if !g.isOCS {
		return
	}

	bounds := screen.Bounds()
	// Position below the timer (timer is at Y=20, so position OCS at Y=50)
	ocsY := 50
	ocsX := bounds.Dx()/2 - 40 // Center horizontally
	ocsWidth := 80
	ocsHeight := 15

	// Create red rectangle
	redRect := ebiten.NewImage(ocsWidth, ocsHeight)
	redRect.Fill(color.RGBA{255, 0, 0, 255}) // Red background

	// Draw red rectangle
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(ocsX), float64(ocsY))
	screen.DrawImage(redRect, op)

	// Draw white text on red background
	ebitenutil.DebugPrintAt(screen, "*** OCS ***", ocsX, ocsY)
}

// isWithinLineBounds checks if the boat's bow position is within the start/finish line bounds
// (between pin and committee boat)
func (g *GameState) isWithinLineBounds(bowPos geometry.Point) bool {
	lineStart := g.Dashboard.LineStart
	lineEnd := g.Dashboard.LineEnd

	// Check if X coordinate is between pin and committee boat
	minX := math.Min(lineStart.X, lineEnd.X)
	maxX := math.Max(lineStart.X, lineEnd.X)

	return bowPos.X >= minX && bowPos.X <= maxX
}

// updateMarkRounding tracks the three phases of mark rounding
func (g *GameState) updateMarkRounding() {
	// Get upwind mark position (it's the third mark in the arena)
	if len(g.Arena.Marks) < 3 {
		return
	}
	upwindMark := g.Arena.Marks[2] // Upwind mark

	boatPos := g.Boat.Pos

	// Phase 1: Sailed past mark (south to north of mark)
	if !g.markRoundingPhase1 {
		// Check if boat has moved from south (Y > markY) to north (Y < markY) of mark
		// We use a 1 unit difference as specified
		if boatPos.Y <= upwindMark.Pos.Y-1 {
			g.markRoundingPhase1 = true
		}
	}

	// Phase 2: Travelled to left (east to west while north of mark)
	if g.markRoundingPhase1 && !g.markRoundingPhase2 {
		// Only check this phase while boat is north of the mark
		if boatPos.Y < upwindMark.Pos.Y {
			// Check if boat has moved from east (X > markX) to west (X < markX) of mark
			if boatPos.X <= upwindMark.Pos.X-1 {
				g.markRoundingPhase2 = true
			}
		} else {
			// If boat moves back south of mark before completing phase 2, reset phase 2
			// but keep phase 1 completed
			g.markRoundingPhase2 = false
		}
	}

	// Phase 3: Sailed below mark (north to south of mark)
	if g.markRoundingPhase1 && g.markRoundingPhase2 && !g.markRoundingPhase3 {
		// Check if boat has moved from north (Y < markY) to south (Y > markY) of mark
		if boatPos.Y >= upwindMark.Pos.Y+1 {
			g.markRoundingPhase3 = true
			g.markRounded = true // All phases complete
		}
	}

	// Reset phase 2 if boat drifts back to east while still north of mark
	if g.markRoundingPhase2 && !g.markRoundingPhase3 && boatPos.Y < upwindMark.Pos.Y {
		if boatPos.X > upwindMark.Pos.X {
			g.markRoundingPhase2 = false
		}
	}
}

// checkFinishLineCrossing detects when boat crosses finish line from course side
func (g *GameState) checkFinishLineCrossing() {
	// Finish line is same as starting line
	startLineY := 2400.0
	bowPos := g.Boat.GetBowPosition()

	// Check if bow crosses the Y coordinate from above (prevBowPos.Y < startLineY) to below (bowPos.Y >= startLineY)
	// AND the boat is within line bounds at the moment of crossing
	// Boat must be coming from course side (north) and cross to finish side (south) while between pin and committee boat
	if g.prevBowPos.Y < startLineY && bowPos.Y >= startLineY && g.isWithinLineBounds(bowPos) {
		// Boat has finished the race!
		g.raceFinished = true
		g.finishTime = g.raceTimer
		g.showFinishBanner = true
		g.finishBannerTime = time.Now()
	}
}

// drawFinishBanner displays the RACE FINISHED banner with race time
func (g *GameState) drawFinishBanner(screen *ebiten.Image) {
	bounds := screen.Bounds()

	// Semi-transparent overlay using vector drawing
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 100}, false)

	// Calculate finish time in minutes and seconds
	minutes := int(g.finishTime.Minutes())
	seconds := int(g.finishTime.Seconds()) % 60
	centiseconds := int((g.finishTime.Milliseconds() % 1000) / 10)

	// FINISH banner text with race time
	finishText := fmt.Sprintf("*** RACE FINISHED! ***\nTime: %02d:%02d.%02d", minutes, seconds, centiseconds)

	// Center the text
	x := bounds.Dx()/2 - 100 // Approximate centering (wider than other banners)
	y := bounds.Dy()/2 - 30

	ebitenutil.DebugPrintAt(screen, finishText, x, y)
}

func (g *GameState) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}
