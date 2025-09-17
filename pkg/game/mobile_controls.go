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

	// Testing
	showControlsOverride bool // Force show controls on desktop for testing
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

	// Process input if touch detected OR if override enabled for testing
	if !mc.hasTouchInput && !mc.showControlsOverride {
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

// ToggleControlsOverride toggles the display of mobile controls on desktop for testing
func (mc *MobileControls) ToggleControlsOverride() {
	mc.showControlsOverride = !mc.showControlsOverride
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
	// Show controls if touch input detected OR if override enabled for testing
	if !mc.hasTouchInput && !mc.showControlsOverride {
		return
	}

	// Draw left arrow button as red polygon
	leftColor := color.RGBA{200, 50, 50, 200} // Lighter red
	if mc.leftPressed {
		leftColor = color.RGBA{255, 80, 80, 220} // Brighter lighter red when pressed
	}
	mc.drawLeftArrow(screen, mc.leftButton, leftColor)

	// Draw right arrow button as green polygon
	rightColor := color.RGBA{0, 150, 0, 200} // Green
	if mc.rightPressed {
		rightColor = color.RGBA{0, 200, 0, 220} // Brighter green when pressed
	}
	mc.drawRightArrow(screen, mc.rightButton, rightColor)

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
		for i, touchID := range touchIDs {
			x, y := ebiten.TouchPosition(touchID)
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Touch %d: %d,%d", i, x, y), 10, 50+i*15)
		}
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("L:%d,%d R:%d,%d P:%d,%d",
		mc.leftButton.X, mc.leftButton.Y,
		mc.rightButton.X, mc.rightButton.Y,
		mc.pauseButton.X, mc.pauseButton.Y), 10, 120)

	// Debug: Show screen vs logical size
	windowW, windowH := ebiten.WindowSize()
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Window: %dx%d Screen: %dx%d", windowW, windowH, ScreenWidth, ScreenHeight), 10, 140)

	// Debug: Show button press states and touch zones
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Pressed: L:%t R:%t P:%t Override:%t",
		mc.leftPressed, mc.rightPressed, mc.pausePressed, mc.showControlsOverride), 10, 160)

	// Debug: Show if any touches are in button areas
	if len(touchIDs) > 0 {
		touchID := touchIDs[0]
		x, y := ebiten.TouchPosition(touchID)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Touch in L:%t R:%t P:%t",
			mc.leftButton.Contains(x, y), mc.rightButton.Contains(x, y), mc.pauseButton.Contains(x, y)), 10, 180)
	}
}

// drawLeftArrow draws a left-pointing arrow polygon
func (mc *MobileControls) drawLeftArrow(screen *ebiten.Image, zone TouchZone, fillColor color.RGBA) {
	if !zone.Enabled {
		return
	}

	// Calculate arrow dimensions within the button zone
	margin := float32(zone.Width) * 0.2 // Margin from button edges
	centerX := float32(zone.X + zone.Width/2)
	centerY := float32(zone.Y + zone.Height/2)

	// Arrow dimensions
	arrowWidth := float32(zone.Width) - 2*margin
	arrowHeight := float32(zone.Height) - 2*margin

	// Draw arrow as combination of rectangles
	// Arrow shaft (horizontal rectangle)
	shaftWidth := arrowWidth * 0.6
	shaftHeight := arrowHeight * 0.3
	shaftX := centerX + arrowWidth/2 - shaftWidth // Position shaft on the right for left arrow
	shaftY := centerY - shaftHeight/2

	vector.DrawFilledRect(screen, shaftX, shaftY, shaftWidth, shaftHeight, fillColor, false)

	// Arrow head (proper triangle pointing left)
	headEndX := shaftX // Triangle ends where shaft begins
	headWidth := arrowWidth - shaftWidth
	headHeight := arrowHeight

	// Draw triangle head by drawing horizontal lines from center outward
	centerLine := int(centerY)
	halfHeight := int(headHeight / 2)

	for i := 0; i <= halfHeight; i++ {
		// Calculate width at this height (linear decrease from base to tip)
		lineWidth := headWidth * (1 - float32(i)/float32(halfHeight))

		if lineWidth > 1 {
			// For left arrow, draw from the tip position (headEndX - headWidth) forward
			startX := headEndX - lineWidth

			// Draw line above center
			if centerLine-i >= int(centerY-headHeight/2) {
				vector.DrawFilledRect(screen, startX, float32(centerLine-i), lineWidth, 1, fillColor, false)
			}
			// Draw line below center (skip center line to avoid double drawing)
			if i > 0 && centerLine+i <= int(centerY+headHeight/2) {
				vector.DrawFilledRect(screen, startX, float32(centerLine+i), lineWidth, 1, fillColor, false)
			}
		}
	}
}

// drawRightArrow draws a right-pointing arrow polygon
func (mc *MobileControls) drawRightArrow(screen *ebiten.Image, zone TouchZone, fillColor color.RGBA) {
	if !zone.Enabled {
		return
	}

	// Calculate arrow dimensions within the button zone
	margin := float32(zone.Width) * 0.2 // Margin from button edges
	centerX := float32(zone.X + zone.Width/2)
	centerY := float32(zone.Y + zone.Height/2)

	// Arrow dimensions
	arrowWidth := float32(zone.Width) - 2*margin
	arrowHeight := float32(zone.Height) - 2*margin

	// Draw arrow as combination of rectangles
	// Arrow shaft (horizontal rectangle)
	shaftWidth := arrowWidth * 0.6
	shaftHeight := arrowHeight * 0.3
	shaftX := centerX - arrowWidth/2
	shaftY := centerY - shaftHeight/2

	vector.DrawFilledRect(screen, shaftX, shaftY, shaftWidth, shaftHeight, fillColor, false)

	// Arrow head (proper triangle pointing right)
	headStartX := shaftX + shaftWidth
	headWidth := arrowWidth - shaftWidth
	headHeight := arrowHeight

	// Draw triangle head by drawing horizontal lines from center outward
	centerLine := int(centerY)
	halfHeight := int(headHeight / 2)

	for i := 0; i <= halfHeight; i++ {
		// Calculate width at this height (linear decrease from base to tip)
		lineWidth := headWidth * (1 - float32(i)/float32(halfHeight))

		if lineWidth > 1 {
			// Draw line above center
			if centerLine-i >= int(centerY-headHeight/2) {
				vector.DrawFilledRect(screen, headStartX, float32(centerLine-i), lineWidth, 1, fillColor, false)
			}
			// Draw line below center (skip center line to avoid double drawing)
			if i > 0 && centerLine+i <= int(centerY+headHeight/2) {
				vector.DrawFilledRect(screen, headStartX, float32(centerLine+i), lineWidth, 1, fillColor, false)
			}
		}
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
