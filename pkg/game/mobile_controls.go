package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// MobileControls handles touch-based input for mobile devices
type MobileControls struct {
	// Touch button zones
	leftButton  TouchZone
	rightButton TouchZone
	pauseButton TouchZone

	// Additional UI button zones (for menu functionality)
	menuButton    TouchZone
	restartButton TouchZone
	timerButton   TouchZone

	// Button press states
	leftPressed    bool
	rightPressed   bool
	pausePressed   bool
	restartPressed bool
	timerPressed   bool

	// State
	menuOpen      bool
	lastTouchTime int
	hasTouchInput bool // Track if we've ever seen touch input
}

// TouchZone defines a rectangular touch area
type TouchZone struct {
	X, Y, Width, Height int
	Enabled             bool
}

// NewMobileControls creates a new mobile controls instance
func NewMobileControls(screenWidth, screenHeight int) *MobileControls {
	buttonSize := 80
	margin := 20

	return &MobileControls{
		// Left arrow button in lower left corner
		leftButton: TouchZone{
			X: margin, Y: screenHeight - buttonSize - margin,
			Width: buttonSize, Height: buttonSize,
			Enabled: true,
		},
		// Right arrow button in lower right corner
		rightButton: TouchZone{
			X: screenWidth - buttonSize - margin, Y: screenHeight - buttonSize - margin,
			Width: buttonSize, Height: buttonSize,
			Enabled: true,
		},
		// Pause/play button in center bottom
		pauseButton: TouchZone{
			X: screenWidth/2 - buttonSize/2, Y: screenHeight - buttonSize - margin,
			Width: buttonSize, Height: buttonSize,
			Enabled: true,
		},

		// Menu button in top left corner (for additional options)
		menuButton: TouchZone{
			X: margin, Y: margin,
			Width: buttonSize / 2, Height: buttonSize / 2,
			Enabled: true,
		},
		restartButton: TouchZone{
			X: margin, Y: margin + buttonSize/2 + 10,
			Width: buttonSize / 2, Height: buttonSize / 2,
			Enabled: false, // Only shown when menu is open
		},
		timerButton: TouchZone{
			X: margin, Y: margin + buttonSize + 20,
			Width: buttonSize / 2, Height: buttonSize / 2,
			Enabled: false, // Only shown when menu is open
		},
	}
}

// Contains checks if a point is within the touch zone
func (tz *TouchZone) Contains(x, y int) bool {
	return tz.Enabled &&
		x >= tz.X && x < tz.X+tz.Width &&
		y >= tz.Y && y < tz.Y+tz.Height
}

// Update processes touch input for mobile controls
func (mc *MobileControls) Update() {
	// Reset button press states
	mc.leftPressed = false
	mc.rightPressed = false
	mc.pausePressed = false
	mc.restartPressed = false
	mc.timerPressed = false

	// Check for any touch input to determine if this is a mobile device
	touchIDs := ebiten.AppendTouchIDs(nil)
	if len(touchIDs) > 0 && !mc.hasTouchInput {
		mc.hasTouchInput = true
	}

	if !mc.hasTouchInput {
		return
	}

	// Get all current touches (including held touches)
	currentTouchIDs := ebiten.AppendTouchIDs(nil)

	// Check each button for current touches (held down)
	for _, id := range currentTouchIDs {
		x, y := ebiten.TouchPosition(id)

		if mc.leftButton.Contains(x, y) {
			mc.leftPressed = true
		}
		if mc.rightButton.Contains(x, y) {
			mc.rightPressed = true
		}
	}

	// Get just pressed touches for one-time button interactions (pause, menu, etc.)
	justPressedTouchIDs := inpututil.AppendJustPressedTouchIDs(nil)

	// Check menu and action buttons for just pressed touches
	for _, id := range justPressedTouchIDs {
		x, y := ebiten.TouchPosition(id)

		if mc.pauseButton.Contains(x, y) {
			mc.pausePressed = true
		}
		if mc.menuButton.Contains(x, y) {
			mc.menuOpen = !mc.menuOpen
			mc.updateMenuButtons()
		}
		if mc.restartButton.Contains(x, y) && mc.restartButton.Enabled {
			mc.restartPressed = true
		}
		if mc.timerButton.Contains(x, y) && mc.timerButton.Enabled {
			mc.timerPressed = true
		}
	}
}

// GetMobileInput returns the current mobile input state
func (mc *MobileControls) GetMobileInput() MobileInput {
	return MobileInput{
		TurnLeft:         mc.leftPressed,
		TurnRight:        mc.rightPressed,
		PausePressed:     mc.pausePressed,
		RestartPressed:   mc.restartButton.Enabled && mc.restartPressed,
		TimerJumpPressed: mc.timerButton.Enabled && mc.timerPressed,
	}
}

// updateMenuButtons toggles visibility of menu-related buttons
func (mc *MobileControls) updateMenuButtons() {
	mc.restartButton.Enabled = mc.menuOpen
	mc.timerButton.Enabled = mc.menuOpen
}

// MobileInput represents the current mobile input state
type MobileInput struct {
	TurnLeft         bool
	TurnRight        bool
	PausePressed     bool
	RestartPressed   bool
	TimerJumpPressed bool
}

// Draw renders the mobile control elements on screen
func (mc *MobileControls) Draw(screen *ebiten.Image, isPaused bool) {
	// Only show controls if we've detected touch input (actual mobile device)
	if !mc.hasTouchInput {
		return
	}

	// Draw left arrow button
	leftColor := color.RGBA{100, 100, 100, 200}
	if mc.leftPressed {
		leftColor = color.RGBA{150, 150, 150, 220} // Highlighted when pressed
	}
	mc.drawButton(screen, mc.leftButton, "◀", leftColor)

	// Draw right arrow button
	rightColor := color.RGBA{100, 100, 100, 200}
	if mc.rightPressed {
		rightColor = color.RGBA{150, 150, 150, 220} // Highlighted when pressed
	}
	mc.drawButton(screen, mc.rightButton, "▶", rightColor)

	// Draw pause/play button in center
	pauseColor := color.RGBA{120, 120, 120, 200}
	if mc.pausePressed {
		pauseColor = color.RGBA{170, 170, 170, 220} // Highlighted when pressed
	}
	pauseText := "||" // Pause symbol
	if isPaused {
		pauseText = ">" // Play symbol
	}
	mc.drawButton(screen, mc.pauseButton, pauseText, pauseColor)

	// Draw smaller menu button in top left
	mc.drawButton(screen, mc.menuButton, "☰", color.RGBA{80, 80, 80, 200})

	// Draw menu buttons if menu is open
	if mc.menuOpen {
		mc.drawButton(screen, mc.restartButton, "↻", color.RGBA{150, 100, 100, 200})
		mc.drawButton(screen, mc.timerButton, "+10", color.RGBA{100, 150, 100, 200})
	}

	// Debug: Show button positions and current touches
	touchIDs := ebiten.AppendTouchIDs(nil)
	if len(touchIDs) > 0 {
		for _, touchID := range touchIDs {
			x, y := ebiten.TouchPosition(touchID)
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Touch: %d,%d", x, y), 10, 50)
		}
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("L:%d,%d R:%d,%d P:%d,%d",
		mc.leftButton.X, mc.leftButton.Y,
		mc.rightButton.X, mc.rightButton.Y,
		mc.pauseButton.X, mc.pauseButton.Y), 10, 70)

	// Debug: Show button press states
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Pressed: L:%t R:%t P:%t",
		mc.leftPressed, mc.rightPressed, mc.pausePressed), 10, 90)
}

// drawButton draws a simple button with text
func (mc *MobileControls) drawButton(screen *ebiten.Image, zone TouchZone, text string, bg color.RGBA) {
	if !zone.Enabled {
		return
	}

	// Draw button background
	vector.DrawFilledRect(screen,
		float32(zone.X), float32(zone.Y),
		float32(zone.Width), float32(zone.Height),
		bg, false)

	// Draw button border
	vector.StrokeRect(screen,
		float32(zone.X), float32(zone.Y),
		float32(zone.Width), float32(zone.Height),
		2, color.RGBA{255, 255, 255, 150}, false)

	// Draw button text (centered)
	ebitenutil.DebugPrintAt(screen, text, zone.X+zone.Width/2-10, zone.Y+zone.Height/2-6)
}
