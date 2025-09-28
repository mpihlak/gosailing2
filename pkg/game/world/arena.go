package world

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/mpihlak/gosailing2/pkg/geometry"
)

type Mark struct {
	Pos  geometry.Point
	Name string
}

func (m *Mark) Draw(screen *ebiten.Image) {
	if m.Name == "Pin" {
		// Draw a small flag at the pin end
		// Flag pole (vertical line)
		ebitenutil.DrawLine(screen, m.Pos.X, m.Pos.Y-10, m.Pos.X, m.Pos.Y+5, color.RGBA{139, 69, 19, 255}) // Brown pole
		// Flag (small triangle)
		flagColor := color.RGBA{255, 0, 0, 255} // Red flag
		// Draw flag as small filled triangle
		for i := 0; i < 6; i++ {
			ebitenutil.DrawLine(screen, m.Pos.X, m.Pos.Y-10+float64(i), m.Pos.X+8-float64(i), m.Pos.Y-10+float64(i), flagColor)
		}
		// Mark base (small circle)
		ebitenutil.DrawRect(screen, m.Pos.X-2, m.Pos.Y-2, 4, 4, color.RGBA{255, 0, 0, 255})
	} else if m.Name == "Upwind" {
		// Draw upwind mark with orange flag (same design as pin)
		// Flag pole (vertical line)
		ebitenutil.DrawLine(screen, m.Pos.X, m.Pos.Y-10, m.Pos.X, m.Pos.Y+5, color.RGBA{139, 69, 19, 255}) // Brown pole
		// Flag (small triangle)
		flagColor := color.RGBA{255, 165, 0, 255} // Orange flag
		// Draw flag as small filled triangle
		for i := 0; i < 6; i++ {
			ebitenutil.DrawLine(screen, m.Pos.X, m.Pos.Y-10+float64(i), m.Pos.X+8-float64(i), m.Pos.Y-10+float64(i), flagColor)
		}
		// Mark base (small circle)
		ebitenutil.DrawRect(screen, m.Pos.X-2, m.Pos.Y-2, 4, 4, color.RGBA{255, 165, 0, 255}) // Orange base
	} else {
		// Draw regular mark (committee boat)
		ebitenutil.DrawRect(screen, m.Pos.X-5, m.Pos.Y-5, 10, 10, color.RGBA{255, 0, 0, 255})
	}
}

type Arena struct {
	Marks []*Mark
}

// drawDottedLine draws a dotted line between two points
func (a *Arena) drawDottedLine(screen *ebiten.Image, x1, y1, x2, y2 float64, lineColor color.Color) {
	dx := x2 - x1
	dy := y2 - y1
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance == 0 {
		return
	}

	// Normalize direction vector
	unitX := dx / distance
	unitY := dy / distance

	// Draw dotted line with 5 pixel segments and 2.5 pixel gaps
	segmentLength := 5.0
	gapLength := 2.5
	totalStep := segmentLength + gapLength

	for t := 0.0; t < distance; t += totalStep {
		// Start of segment
		startX := x1 + unitX*t
		startY := y1 + unitY*t

		// End of segment (don't go beyond the line end)
		endT := math.Min(t+segmentLength, distance)
		endX := x1 + unitX*endT
		endY := y1 + unitY*endT

		// Draw the segment
		ebitenutil.DrawLine(screen, startX, startY, endX, endY, lineColor)
	}
}

// drawWindBarb draws a wind barb at the specified position showing wind direction and strength
func (a *Arena) drawWindBarb(screen *ebiten.Image, x, y float64, windDir, windSpeed float64) {
	// Light gray color as requested
	windColor := color.RGBA{192, 192, 192, 255}

	// Wind barb shaft length (main line showing direction)
	shaftLength := 20.0

	// Convert wind direction to radians (wind direction is where wind comes FROM)
	dirRad := windDir * math.Pi / 180.0

	// Calculate shaft end point - shaft points in direction wind is blowing TO
	shaftEndX := x + shaftLength*math.Sin(dirRad+math.Pi)
	shaftEndY := y - shaftLength*math.Cos(dirRad+math.Pi)

	// Draw main shaft
	ebitenutil.DrawLine(screen, x, y, shaftEndX, shaftEndY, windColor)

	// Draw wind speed indicators (barbs/flags)
	// Each full barb represents 10 knots, half barbs represent 5 knots
	fullBarbs := int(windSpeed / 10)
	halfBarb := (int(windSpeed) % 10) >= 5

	// Barb length and perpendicular angle
	barbLength := 8.0
	perpAngle := (dirRad + math.Pi) + math.Pi/2 // Perpendicular to shaft direction

	// Draw full barbs (every 10 knots)
	for i := 0; i < fullBarbs && i < 5; i++ { // Limit to 5 barbs to keep it clean
		// Position along shaft (starting from base, moving toward end)
		barbPos := 0.2 + float64(i)*0.15
		if barbPos > 0.8 {
			barbPos = 0.8
		}

		barbStartX := x + barbPos*shaftLength*math.Sin(dirRad+math.Pi)
		barbStartY := y - barbPos*shaftLength*math.Cos(dirRad+math.Pi)
		barbEndX := barbStartX + barbLength*math.Sin(perpAngle)
		barbEndY := barbStartY - barbLength*math.Cos(perpAngle)

		ebitenutil.DrawLine(screen, barbStartX, barbStartY, barbEndX, barbEndY, windColor)
	}

	// Draw half barb if needed (5 knots)
	if halfBarb {
		barbPos := 0.2 + float64(fullBarbs)*0.15
		if barbPos > 0.8 {
			barbPos = 0.8
		}

		barbStartX := x + barbPos*shaftLength*math.Sin(dirRad+math.Pi)
		barbStartY := y - barbPos*shaftLength*math.Cos(dirRad+math.Pi)
		barbEndX := barbStartX + (barbLength*0.5)*math.Sin(perpAngle)
		barbEndY := barbStartY - (barbLength*0.5)*math.Cos(perpAngle)

		ebitenutil.DrawLine(screen, barbStartX, barbStartY, barbEndX, barbEndY, windColor)
	}
}

// drawLaylines draws the starboard and port laylines for the upwind mark
func (a *Arena) drawLaylines(screen *ebiten.Image) {
	// Find upwind mark (third mark in the array)
	if len(a.Marks) < 3 {
		return
	}
	upwindMark := a.Marks[2]

	// Wind is from North (0 degrees), laylines show the close-hauled approach paths to the mark
	// Since positive Y is down (toward starting line), we want laylines extending in positive Y direction
	// Starboard tack: boats sail at 45° to wind (northeast), layline extends southwest from mark
	// Port tack: boats sail at -45° to wind (northwest), layline extends southeast from mark

	laylineColor := color.RGBA{128, 128, 128, 100} // Light gray with transparency

	// Calculate layline length (extend toward starting line)
	laylineLength := 1500.0

	// Starboard layline: extends southwest from mark (225°)
	starboardAngle := 225.0 * math.Pi / 180
	starboardEndX := upwindMark.Pos.X + laylineLength*math.Sin(starboardAngle)
	starboardEndY := upwindMark.Pos.Y - laylineLength*math.Cos(starboardAngle) // Negative cos(225°) makes this positive Y

	// Port layline: extends southeast from mark (135°)
	portAngle := 135.0 * math.Pi / 180
	portEndX := upwindMark.Pos.X + laylineLength*math.Sin(portAngle)
	portEndY := upwindMark.Pos.Y - laylineLength*math.Cos(portAngle) // Negative cos(135°) makes this positive Y

	// Draw both laylines as dotted lines (extending toward starting line)
	a.drawDottedLine(screen, upwindMark.Pos.X, upwindMark.Pos.Y, starboardEndX, starboardEndY, laylineColor)
	a.drawDottedLine(screen, upwindMark.Pos.X, upwindMark.Pos.Y, portEndX, portEndY, laylineColor)
}

// drawWindIndicators draws wind barbs across the course at regular intervals
func (a *Arena) drawWindIndicators(screen *ebiten.Image, wind Wind) {
	// Grid spacing - every 150 pixels as requested
	gridSpacing := 150.0

	// Get screen bounds to know the area we need to cover
	bounds := screen.Bounds()
	startX := 0.0
	startY := 0.0
	endX := float64(bounds.Max.X)
	endY := float64(bounds.Max.Y)

	// Draw wind barbs at grid points
	for x := startX; x <= endX; x += gridSpacing {
		for y := startY; y <= endY; y += gridSpacing {
			// Get wind at this position
			windDir, windSpeed := wind.GetWind(geometry.Point{X: x, Y: y})

			// Draw wind barb at this grid point
			a.drawWindBarb(screen, x, y, windDir, windSpeed)
		}
	}
}

func (a *Arena) Draw(screen *ebiten.Image, raceStarted bool, wind Wind) {
	// Draw wind indicators first (in background)
	if wind != nil {
		a.drawWindIndicators(screen, wind)
	}

	// Draw starting line if we have exactly 2 marks (Pin and Committee)
	if len(a.Marks) == 2 {
		pin := a.Marks[0]
		committee := a.Marks[1]

		// Choose line color based on race state
		var lineColor color.Color
		if raceStarted {
			lineColor = color.RGBA{0, 255, 0, 255} // Green when race started
		} else {
			lineColor = color.RGBA{255, 255, 255, 255} // White before start
		}

		// Draw dotted line
		a.drawDottedLine(screen, pin.Pos.X, pin.Pos.Y, committee.Pos.X, committee.Pos.Y, lineColor)
	}

	// Draw laylines for upwind mark (if we have 3 marks including upwind)
	if len(a.Marks) >= 3 {
		a.drawLaylines(screen)
	}

	// Draw marks
	for _, mark := range a.Marks {
		mark.Draw(screen)
	}
}
