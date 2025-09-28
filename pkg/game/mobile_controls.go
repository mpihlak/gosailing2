package game

import (
	"fmt"
	"image/color"
	"math"

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

	// Additional UI button zones
	restartButton TouchZone

	// Button press states
	leftPressed    bool
	rightPressed   bool
	pausePressed   bool
	restartPressed bool

	// State
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

	mc := &MobileControls{
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

		// Restart button in top left corner
		restartButton: TouchZone{
			X: margin, Y: margin,
			Width: buttonSize * 2 / 3, Height: buttonSize * 2 / 3, // Slightly larger than old menu button
			Enabled: true,
		},
	}

	// Determine touch capability at initialization
	mc.detectTouchCapability()

	return mc
}

// detectTouchCapability determines if the device supports touch input
func (mc *MobileControls) detectTouchCapability() {
	// Check if there are any active touch points
	touchIDs := ebiten.AppendTouchIDs(nil)
	if len(touchIDs) > 0 {
		mc.hasTouchInput = true
		return
	}

	// Start with no touch input detected
	// This will be updated dynamically during Update() when touch is first detected
	mc.hasTouchInput = false
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

	// Dynamically detect touch input during runtime
	// Check both current touches and just-pressed touches
	touchIDs := ebiten.AppendTouchIDs(nil)
	justPressed := inpututil.AppendJustPressedTouchIDs(nil)
	if len(touchIDs) > 0 || len(justPressed) > 0 {
		mc.hasTouchInput = true
	}

	// Process input if touch detected OR if override enabled for testing
	if !mc.hasTouchInput && !mc.showControlsOverride {
		return
	}

	// Get all current touches (including held touches)
	currentTouchIDs := touchIDs

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

	// Check action buttons for just pressed touches
	for _, id := range justPressedTouchIDs {
		x, y := ebiten.TouchPosition(id)

		if mc.pauseButton.Contains(x, y) {
			mc.pausePressed = true
		}
		if mc.restartButton.Contains(x, y) {
			mc.restartPressed = true
		}
	}
}

// GetMobileInput returns the current mobile input state
func (mc *MobileControls) GetMobileInput() MobileInput {
	return MobileInput{
		TurnLeft:       mc.leftPressed,
		TurnRight:      mc.rightPressed,
		PausePressed:   mc.pausePressed,
		RestartPressed: mc.restartPressed,
	}
}

// ToggleControlsOverride toggles the display of mobile controls on desktop for testing
func (mc *MobileControls) ToggleControlsOverride() {
	mc.showControlsOverride = !mc.showControlsOverride
}

// MobileInput represents the current mobile input state
type MobileInput struct {
	TurnLeft       bool
	TurnRight      bool
	PausePressed   bool
	RestartPressed bool
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

	// Draw pause/play button in center as polygon
	pauseColor := color.RGBA{120, 120, 120, 200}
	if mc.pausePressed {
		pauseColor = color.RGBA{170, 170, 170, 220} // Highlighted when pressed
	}
	if isPaused {
		// Game is paused - show green play triangle (pointing right)
		playColor := color.RGBA{0, 150, 0, 200}
		if mc.pausePressed {
			playColor = color.RGBA{0, 200, 0, 220}
		}
		mc.drawPlayTriangle(screen, mc.pauseButton, playColor)
	} else {
		// Game is running - show pause bars (two vertical rectangles)
		mc.drawPauseBars(screen, mc.pauseButton, pauseColor)
	}

	// Draw restart button in top left as curved arrow
	restartColor := color.RGBA{80, 120, 200, 200} // Blue
	if mc.restartPressed {
		restartColor = color.RGBA{120, 160, 255, 220} // Brighter blue when pressed
	}
	mc.drawRestartArrow(screen, mc.restartButton, restartColor)

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

// drawRestartArrow draws a curved retry arrow
func (mc *MobileControls) drawRestartArrow(screen *ebiten.Image, zone TouchZone, fillColor color.RGBA) {
	if !zone.Enabled {
		return
	}

	// Calculate arrow dimensions within the button zone
	margin := float32(zone.Width) * 0.15
	centerX := float32(zone.X + zone.Width/2)
	centerY := float32(zone.Y + zone.Height/2)

	// Arrow dimensions
	arrowSize := float32(zone.Width) - 2*margin
	radius := arrowSize * 0.35

	// Draw curved arrow as multiple segments
	// We'll approximate the curve using small rectangular segments
	strokeWidth := float32(3)

	// Draw the main arc (about 240 degrees) with larger gap
	startAngle := float32(0.8) // Start further around from right side
	endAngle := float32(6.0)   // End earlier to create bigger gap (240 degrees)
	steps := 30

	for i := 0; i < steps; i++ {
		angle1 := startAngle + (endAngle-startAngle)*float32(i)/float32(steps)
		angle2 := startAngle + (endAngle-startAngle)*float32(i+1)/float32(steps)

		// Calculate points on the circle
		x1 := centerX + radius*float32(math.Cos(float64(angle1)))
		y1 := centerY + radius*float32(math.Sin(float64(angle1)))
		x2 := centerX + radius*float32(math.Cos(float64(angle2)))
		y2 := centerY + radius*float32(math.Sin(float64(angle2)))

		// Draw line segment
		vector.StrokeLine(screen, x1, y1, x2, y2, strokeWidth, fillColor, false)
	}

	// Draw arrowhead at the end of the arc
	// Calculate the end position and direction
	endX := centerX + radius*float32(math.Cos(float64(endAngle)))
	endY := centerY + radius*float32(math.Sin(float64(endAngle)))

	// Arrowhead size
	arrowHeadSize := arrowSize * 0.2

	// Direction perpendicular to the arc (tangent direction)
	tangentAngle := endAngle + 1.57 // Add 90 degrees for tangent
	tangentX := float32(math.Cos(float64(tangentAngle)))
	tangentY := float32(math.Sin(float64(tangentAngle)))

	// Perpendicular direction for arrowhead width
	perpX := -tangentY
	perpY := tangentX

	// Draw arrowhead as small filled triangle
	headTipX := endX + tangentX*arrowHeadSize
	headTipY := endY + tangentY*arrowHeadSize
	headBase1X := endX + perpX*arrowHeadSize*0.5
	headBase1Y := endY + perpY*arrowHeadSize*0.5
	headBase2X := endX - perpX*arrowHeadSize*0.5
	headBase2Y := endY - perpY*arrowHeadSize*0.5

	// Draw arrowhead using lines
	vector.StrokeLine(screen, headTipX, headTipY, headBase1X, headBase1Y, strokeWidth, fillColor, false)
	vector.StrokeLine(screen, headTipX, headTipY, headBase2X, headBase2Y, strokeWidth, fillColor, false)
	vector.StrokeLine(screen, headBase1X, headBase1Y, headBase2X, headBase2Y, strokeWidth, fillColor, false)
}

// drawPlayTriangle draws a right-pointing play triangle
func (mc *MobileControls) drawPlayTriangle(screen *ebiten.Image, zone TouchZone, fillColor color.RGBA) {
	if !zone.Enabled {
		return
	}

	// Calculate triangle dimensions within the button zone
	margin := float32(zone.Width) * 0.25 // Larger margin for triangle
	centerX := float32(zone.X + zone.Width/2)
	centerY := float32(zone.Y + zone.Height/2)

	// Triangle dimensions
	triangleWidth := float32(zone.Width) - 2*margin
	triangleHeight := float32(zone.Height) - 2*margin

	// Triangle points (right-pointing)
	leftX := centerX - triangleWidth/2
	topY := centerY - triangleHeight/2
	bottomY := centerY + triangleHeight/2

	// Draw triangle using horizontal lines from center outward
	centerLine := int(centerY)
	halfHeight := int(triangleHeight / 2)

	for i := 0; i <= halfHeight; i++ {
		// Calculate width at this height (linear decrease from base to tip)
		lineWidth := triangleWidth * (1 - float32(i)/float32(halfHeight))

		if lineWidth > 1 {
			// Draw line above center
			if centerLine-i >= int(topY) {
				vector.DrawFilledRect(screen, leftX, float32(centerLine-i), lineWidth, 1, fillColor, false)
			}
			// Draw line below center (skip center line to avoid double drawing)
			if i > 0 && centerLine+i <= int(bottomY) {
				vector.DrawFilledRect(screen, leftX, float32(centerLine+i), lineWidth, 1, fillColor, false)
			}
		}
	}
}

// drawPauseBars draws two vertical bars for pause symbol
func (mc *MobileControls) drawPauseBars(screen *ebiten.Image, zone TouchZone, fillColor color.RGBA) {
	if !zone.Enabled {
		return
	}

	// Calculate bar dimensions within the button zone
	margin := float32(zone.Width) * 0.3 // Margin from button edges
	centerX := float32(zone.X + zone.Width/2)
	centerY := float32(zone.Y + zone.Height/2)

	// Bar dimensions
	barHeight := float32(zone.Height) - 2*margin
	barWidth := float32(zone.Width) * 0.15  // Each bar is 15% of button width
	barSpacing := float32(zone.Width) * 0.1 // 10% spacing between bars

	// Left bar
	leftBarX := centerX - barSpacing/2 - barWidth
	barY := centerY - barHeight/2
	vector.DrawFilledRect(screen, leftBarX, barY, barWidth, barHeight, fillColor, false)

	// Right bar
	rightBarX := centerX + barSpacing/2
	vector.DrawFilledRect(screen, rightBarX, barY, barWidth, barHeight, fillColor, false)
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
