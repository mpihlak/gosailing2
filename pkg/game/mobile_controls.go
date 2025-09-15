package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// MobileControls handles touch-based input for mobile devices
type MobileControls struct {
	// Touch steering zones
	leftSteerZone  TouchZone
	rightSteerZone TouchZone

	// UI button zones
	pauseButton   TouchZone
	menuButton    TouchZone
	restartButton TouchZone
	timerButton   TouchZone

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
	buttonSize := 60
	margin := 20

	return &MobileControls{
		// Steering zones take up left and right thirds of screen
		leftSteerZone: TouchZone{
			X: 0, Y: 0,
			Width: screenWidth / 3, Height: screenHeight,
			Enabled: true,
		},
		rightSteerZone: TouchZone{
			X: (screenWidth * 2) / 3, Y: 0,
			Width: screenWidth / 3, Height: screenHeight,
			Enabled: true,
		},

		// UI buttons in corners
		pauseButton: TouchZone{
			X: screenWidth - buttonSize - margin, Y: margin,
			Width: buttonSize, Height: buttonSize,
			Enabled: true,
		},
		menuButton: TouchZone{
			X: margin, Y: margin,
			Width: buttonSize, Height: buttonSize,
			Enabled: true,
		},
		restartButton: TouchZone{
			X: margin, Y: margin + buttonSize + 10,
			Width: buttonSize, Height: buttonSize,
			Enabled: false, // Only shown when menu is open
		},
		timerButton: TouchZone{
			X: margin, Y: margin + (buttonSize+10)*2,
			Width: buttonSize, Height: buttonSize,
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

// Update processes touch input and returns control actions
func (mc *MobileControls) Update() MobileInput {
	input := MobileInput{}

	// Get all current touches
	touchIDs := ebiten.AppendTouchIDs(nil)

	// Get just pressed touches for button handling
	justPressedTouchIDs := inpututil.AppendJustPressedTouchIDs(nil)

	// Track if we've ever seen touch input (indicates mobile device)
	if len(touchIDs) > 0 || len(justPressedTouchIDs) > 0 {
		mc.hasTouchInput = true
	}

	// Process just pressed touches for button actions
	for _, touchID := range justPressedTouchIDs {
		x, y := ebiten.TouchPosition(touchID)

		// Handle button presses
		if mc.pauseButton.Contains(x, y) {
			input.PausePressed = true
		} else if mc.menuButton.Contains(x, y) {
			mc.menuOpen = !mc.menuOpen
			mc.updateMenuButtons()
		} else if mc.restartButton.Contains(x, y) && mc.menuOpen {
			input.RestartPressed = true
			mc.menuOpen = false
			mc.updateMenuButtons()
		} else if mc.timerButton.Contains(x, y) && mc.menuOpen {
			input.TimerJumpPressed = true
		}
	}

	// Process all current touches for continuous steering
	for _, touchID := range touchIDs {
		x, y := ebiten.TouchPosition(touchID)

		// Handle continuous steering input
		if mc.leftSteerZone.Contains(x, y) {
			input.TurnLeft = true
		} else if mc.rightSteerZone.Contains(x, y) {
			input.TurnRight = true
		}
	}

	return input
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
func (mc *MobileControls) Draw(screen *ebiten.Image) {
	// Only draw controls if we've detected touch input (mobile device)
	if !mc.hasTouchInput {
		return
	}

	// Get current touches for highlighting active zones
	touchIDs := ebiten.AppendTouchIDs(nil)

	// Draw steering zones (subtle overlay when active)
	for _, touchID := range touchIDs {
		x, y := ebiten.TouchPosition(touchID)

		if mc.leftSteerZone.Contains(x, y) {
			// Draw left steering indicator
			vector.DrawFilledRect(screen,
				float32(mc.leftSteerZone.X), float32(mc.leftSteerZone.Y),
				float32(mc.leftSteerZone.Width), float32(mc.leftSteerZone.Height),
				color.RGBA{255, 255, 255, 30}, false)
			ebitenutil.DebugPrintAt(screen, "◀ TURN LEFT", 20, mc.leftSteerZone.Height/2)
		}

		if mc.rightSteerZone.Contains(x, y) {
			// Draw right steering indicator
			vector.DrawFilledRect(screen,
				float32(mc.rightSteerZone.X), float32(mc.rightSteerZone.Y),
				float32(mc.rightSteerZone.Width), float32(mc.rightSteerZone.Height),
				color.RGBA{255, 255, 255, 30}, false)
			ebitenutil.DebugPrintAt(screen, "TURN RIGHT ▶", mc.rightSteerZone.X+20, mc.rightSteerZone.Height/2)
		}
	}

	// Draw pause button
	mc.drawButton(screen, mc.pauseButton, "⏸", color.RGBA{100, 100, 100, 200})

	// Draw menu button
	mc.drawButton(screen, mc.menuButton, "☰", color.RGBA{100, 100, 100, 200})

	// Draw menu buttons if menu is open
	if mc.menuOpen {
		mc.drawButton(screen, mc.restartButton, "↻", color.RGBA{150, 100, 100, 200})
		mc.drawButton(screen, mc.timerButton, "+10", color.RGBA{100, 150, 100, 200})
	}
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
