package objects

import (
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/mpihlak/ebiten-sailing/pkg/game/world"
	"github.com/mpihlak/ebiten-sailing/pkg/geometry"
	"github.com/mpihlak/ebiten-sailing/pkg/polars"
)

const (
	maxHistoryPoints = 50
	historyInterval  = 200 * time.Millisecond
	boatHeight       = 15.0                  // Triangle height
	boatWidth        = 7.5                   // Triangle width
	speedScale       = 30.0 / 6.0            // Pixels per second per knot (10 pixels/sec at 6 knots)
	boatMass         = 4000.0                // Boat mass in kg
	dragCoefficient  = 0.02                  // Water resistance coefficient (reduced for more gradual deceleration)
	inputDelay       = 30 * time.Millisecond // Delay between keystroke readings
)

type Boat struct {
	Pos         geometry.Point // Center of the boat
	Heading     float64        // in degrees
	Speed       float64        // in knots (current polar speed)
	VelX, VelY  float64        // Actual velocity in pixels/frame
	History     []geometry.Point
	lastHistory time.Time
	lastInput   time.Time     // Last time input was processed
	Polars      polars.Polars // Polar performance data
	Wind        world.Wind    // Wind interface to get wind conditions
}

// GetBowPosition returns the position of the boat's bow (front tip)
func (b *Boat) GetBowPosition() geometry.Point {
	headingRad := b.Heading * math.Pi / 180
	bowDistance := boatHeight / 2

	return geometry.Point{
		X: b.Pos.X + bowDistance*math.Sin(headingRad),
		Y: b.Pos.Y - bowDistance*math.Cos(headingRad),
	}
}

func (b *Boat) Update() {
	// Input handling with delay to prevent overturning
	if time.Since(b.lastInput) >= inputDelay {
		if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
			b.Heading -= 2
			b.lastInput = time.Now()
		}
		if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
			b.Heading += 2
			b.lastInput = time.Now()
		}
	}

	// Normalize heading
	if b.Heading < 0 {
		b.Heading += 360
	}
	if b.Heading >= 360 {
		b.Heading -= 360
	}

	// Get wind conditions at boat position
	windDir, windSpeed := b.Wind.GetWind(b.Pos)

	// Calculate True Wind Angle (TWA)
	twa := b.Heading - windDir
	if twa < -180 {
		twa += 360
	} else if twa > 180 {
		twa -= 360
	}

	// Get target speed from polars
	targetSpeed := b.Polars.GetBoatSpeed(twa, windSpeed)

	// Convert target speed to target velocity in heading direction
	headingRad := b.Heading * math.Pi / 180
	targetPixelSpeed := targetSpeed * speedScale / 60.0
	targetVelX := targetPixelSpeed * math.Sin(headingRad)
	targetVelY := -targetPixelSpeed * math.Cos(headingRad) // Y inverted

	// Calculate current velocity magnitude
	currentSpeed := math.Sqrt(b.VelX*b.VelX + b.VelY*b.VelY)

	// Project current velocity onto the heading direction to maintain forward momentum
	if currentSpeed > 0.01 {
		// Calculate the component of current velocity in the heading direction
		currentHeadingVelX := math.Sin(headingRad)
		currentHeadingVelY := -math.Cos(headingRad)

		// Dot product to get the magnitude of velocity in heading direction
		forwardSpeed := b.VelX*currentHeadingVelX + b.VelY*currentHeadingVelY

		// Keep the forward momentum but gradually align with heading
		alignmentFactor := 0.05 // How quickly the boat aligns velocity with heading
		b.VelX = b.VelX*(1-alignmentFactor) + forwardSpeed*currentHeadingVelX*alignmentFactor
		b.VelY = b.VelY*(1-alignmentFactor) + forwardSpeed*currentHeadingVelY*alignmentFactor
	}

	// Apply drag force (proportional to velocity squared)
	currentSpeed = math.Sqrt(b.VelX*b.VelX + b.VelY*b.VelY)
	dragForce := dragCoefficient * currentSpeed * currentSpeed

	// Calculate drag acceleration (F = ma, so a = F/m)
	dragAccel := dragForce / boatMass * 10 // Reduced scale factor for slower deceleration (was 20)

	// Apply drag in opposite direction of movement
	if currentSpeed > 0.01 { // Avoid division by zero
		dragVelX := -dragAccel * (b.VelX / currentSpeed) / 60.0 // Convert to per-frame
		dragVelY := -dragAccel * (b.VelY / currentSpeed) / 60.0
		b.VelX += dragVelX
		b.VelY += dragVelY
	}

	// Apply force towards target velocity (wind power)
	// This simulates the boat's ability to accelerate towards the polar speed
	accelerationFactor := 0.01 // Reduced for slower acceleration (was 0.02)
	b.VelX += (targetVelX - b.VelX) * accelerationFactor
	b.VelY += (targetVelY - b.VelY) * accelerationFactor

	// Move boat using actual velocity
	b.Pos.X += b.VelX
	b.Pos.Y += b.VelY

	// Calculate actual current speed in knots for dashboard display
	currentPixelSpeed := math.Sqrt(b.VelX*b.VelX + b.VelY*b.VelY)
	b.Speed = currentPixelSpeed * 60.0 / speedScale // Convert back to knots

	// Add to history
	if time.Since(b.lastHistory) >= historyInterval {
		b.History = append(b.History, b.Pos)
		b.lastHistory = time.Now()

		// Cap history at maxHistoryPoints
		if len(b.History) > maxHistoryPoints {
			b.History = b.History[1:]
		}
	}
}

func (b *Boat) Draw(screen *ebiten.Image) {
	// Draw boat history (skip the last 2 points to avoid overlap with boat)
	historyToShow := len(b.History) - 1
	if historyToShow < 0 {
		historyToShow = 0
	}

	for i := 0; i < historyToShow; i++ {
		p := b.History[i]
		ebitenutil.DrawCircle(screen, p.X, p.Y, 2, color.RGBA{173, 216, 230, 150})
	}

	// Draw boat as triangle pointing towards heading
	headingRad := b.Heading * math.Pi / 180

	// Triangle dimensions
	height := boatHeight
	width := boatWidth

	// Calculate triangle vertices relative to boat center position
	// Bow (tip) is forward from center, stern (base) is behind center
	bowDistance := height / 2
	sternDistance := height / 2

	// Bow position (front tip)
	bowX := b.Pos.X + bowDistance*math.Sin(headingRad)
	bowY := b.Pos.Y - bowDistance*math.Cos(headingRad)

	// Stern center position (back center)
	sternX := b.Pos.X - sternDistance*math.Sin(headingRad)
	sternY := b.Pos.Y + sternDistance*math.Cos(headingRad)

	// Left and right stern points
	leftX := sternX - (width/2)*math.Cos(headingRad)
	leftY := sternY - (width/2)*math.Sin(headingRad)

	rightX := sternX + (width/2)*math.Cos(headingRad)
	rightY := sternY + (width/2)*math.Sin(headingRad)

	// Draw triangle using lines
	ebitenutil.DrawLine(screen, bowX, bowY, leftX, leftY, color.White)
	ebitenutil.DrawLine(screen, leftX, leftY, rightX, rightY, color.White)
	ebitenutil.DrawLine(screen, rightX, rightY, bowX, bowY, color.White)
}
