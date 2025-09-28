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

func (a *Arena) Draw(screen *ebiten.Image, raceStarted bool) {
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
