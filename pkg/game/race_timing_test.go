package game

import (
	"testing"
	"time"

	"github.com/mpihlak/gosailing2/pkg/dashboard"
	"github.com/mpihlak/gosailing2/pkg/game/objects"
	"github.com/mpihlak/gosailing2/pkg/game/world"
	"github.com/mpihlak/gosailing2/pkg/geometry"
	"github.com/mpihlak/gosailing2/pkg/polars"
)

// Helper function to create a minimal game state for testing
func createTestGame() *GameState {
	wind := world.NewOscillatingWind(10, 10, WorldWidth)

	pinX := float64(WorldWidth/2 - 200)
	committeeX := float64(WorldWidth/2 + 200)
	lineY := float64(2400)

	boat := &objects.Boat{
		Pos:     geometry.Point{X: 1000, Y: 2500}, // Below starting line
		Heading: 0,                                 // North
		Speed:   6.0,
		Polars:  &polars.RealisticPolar{},
		Wind:    wind,
	}

	arena := &world.Arena{
		Marks: []*world.Mark{
			{Pos: geometry.Point{X: pinX, Y: lineY}, Name: "Pin"},
			{Pos: geometry.Point{X: committeeX, Y: lineY}, Name: "Committee"},
			{Pos: geometry.Point{X: 1000, Y: 1800}, Name: "Upwind"},
		},
	}

	dash := &dashboard.Dashboard{
		Boat:       boat,
		Wind:       wind,
		LineStart:  geometry.Point{X: pinX, Y: lineY},
		LineEnd:    geometry.Point{X: committeeX, Y: lineY},
		UpwindMark: geometry.Point{X: 1000, Y: 1800},
	}

	g := &GameState{
		Boat:           boat,
		Arena:          arena,
		Wind:           wind,
		Dashboard:      dash,
		raceStarted:    false,
		raceFinished:   false,
		timerDuration:  30 * time.Second,
		elapsedTime:    0,
		lastUpdateTime: time.Now(),
		prevBowPos:     boat.GetBowPosition(),
	}

	return g
}

func TestOCS_BoatCrossesLineBeforeStart(t *testing.T) {
	g := createTestGame()
	g.raceStarted = false

	// Place boat's bow above the line (course side)
	g.Boat.Pos = geometry.Point{X: 1000, Y: 2390}
	bowPos := g.Boat.GetBowPosition()

	// Simulate OCS check (from game.go Update logic)
	startLineY := 2400.0
	if bowPos.Y <= startLineY && g.isWithinLineBounds(bowPos) {
		g.isOCS = true
	}

	if !g.isOCS {
		t.Error("Expected boat to be OCS when above line before start")
	}
}

func TestOCS_BoatBelowLineNotOCS(t *testing.T) {
	g := createTestGame()
	g.raceStarted = false

	// Place boat below the line
	g.Boat.Pos = geometry.Point{X: 1000, Y: 2500}
	bowPos := g.Boat.GetBowPosition()

	startLineY := 2400.0
	if bowPos.Y <= startLineY && g.isWithinLineBounds(bowPos) {
		g.isOCS = true
	}

	if g.isOCS {
		t.Error("Expected boat NOT to be OCS when below line")
	}
}

func TestOCS_ClearedByRecrossing(t *testing.T) {
	g := createTestGame()
	g.raceStarted = false
	g.isOCS = true // Boat was OCS

	// Boat recrosses back below the line
	g.Boat.Pos = geometry.Point{X: 1000, Y: 2410}
	bowPos := g.Boat.GetBowPosition()

	startLineY := 2400.0
	if g.isOCS && bowPos.Y > startLineY && g.isWithinLineBounds(bowPos) {
		g.isOCS = false
	}

	if g.isOCS {
		t.Error("Expected OCS to be cleared when boat recrosses line")
	}
}

func TestLineCrossing_AfterRaceStart(t *testing.T) {
	g := createTestGame()
	g.raceStarted = true
	g.isOCS = false
	g.hasCrossedLine = false

	startLineY := 2400.0

	// Set previous bow position below line
	g.prevBowPos = geometry.Point{X: 1000, Y: 2410}

	// Current bow position above line (crossed)
	g.Boat.Pos = geometry.Point{X: 1000, Y: 2390}
	bowPos := g.Boat.GetBowPosition()

	// Simulate line crossing detection
	if g.prevBowPos.Y > startLineY && bowPos.Y <= startLineY && g.isWithinLineBounds(bowPos) {
		g.hasCrossedLine = true
	}

	if !g.hasCrossedLine {
		t.Error("Expected line crossing to be detected after race start")
	}
}

func TestLineCrossing_NotDetectedWhenOCS(t *testing.T) {
	g := createTestGame()
	g.raceStarted = true
	g.isOCS = true // Boat is OCS
	g.hasCrossedLine = false

	startLineY := 2400.0
	g.prevBowPos = geometry.Point{X: 1000, Y: 2410}
	g.Boat.Pos = geometry.Point{X: 1000, Y: 2390}
	bowPos := g.Boat.GetBowPosition()

	// Line crossing should NOT be detected when OCS
	if !g.hasCrossedLine && !g.isOCS {
		if g.prevBowPos.Y > startLineY && bowPos.Y <= startLineY && g.isWithinLineBounds(bowPos) {
			g.hasCrossedLine = true
		}
	}

	if g.hasCrossedLine {
		t.Error("Line crossing should not be detected when boat is OCS")
	}
}

func TestLineCrossing_OutsideLineBounds(t *testing.T) {
	g := createTestGame()
	g.raceStarted = true
	g.isOCS = false
	g.hasCrossedLine = false

	startLineY := 2400.0

	// Set previous position below line, but boat crosses OUTSIDE line bounds
	g.prevBowPos = geometry.Point{X: 500, Y: 2410} // Far left of pin
	g.Boat.Pos = geometry.Point{X: 500, Y: 2390}
	bowPos := g.Boat.GetBowPosition()

	// Should not count as crossing
	if g.prevBowPos.Y > startLineY && bowPos.Y <= startLineY && g.isWithinLineBounds(bowPos) {
		g.hasCrossedLine = true
	}

	if g.hasCrossedLine {
		t.Error("Line crossing should not be detected outside line bounds")
	}
}

func TestFinishLine_DetectedAfterMarkRounded(t *testing.T) {
	g := createTestGame()
	g.raceStarted = true
	g.hasCrossedLine = true
	g.markRounded = true
	g.raceFinished = false

	startLineY := 2400.0

	// Boat approaching from above (course side)
	g.prevBowPos = geometry.Point{X: 1000, Y: 2390}

	// Boat crosses finish line (going south)
	g.Boat.Pos = geometry.Point{X: 1000, Y: 2410}
	bowPos := g.Boat.GetBowPosition()

	// Simulate finish line crossing
	if g.prevBowPos.Y < startLineY && bowPos.Y >= startLineY && g.isWithinLineBounds(bowPos) {
		g.raceFinished = true
	}

	if !g.raceFinished {
		t.Error("Expected race to be finished after crossing finish line")
	}
}

func TestFinishLine_NotDetectedWithoutMarkRounded(t *testing.T) {
	g := createTestGame()
	g.raceStarted = true
	g.hasCrossedLine = true
	g.markRounded = false // Mark NOT rounded
	g.raceFinished = false

	startLineY := 2400.0
	g.prevBowPos = geometry.Point{X: 1000, Y: 2390}
	g.Boat.Pos = geometry.Point{X: 1000, Y: 2410}
	bowPos := g.Boat.GetBowPosition()

	// Should not finish without mark rounded
	if g.hasCrossedLine && g.markRounded {
		if g.prevBowPos.Y < startLineY && bowPos.Y >= startLineY && g.isWithinLineBounds(bowPos) {
			g.raceFinished = true
		}
	}

	if g.raceFinished {
		t.Error("Race should not finish without mark being rounded")
	}
}

func TestIsWithinLineBounds(t *testing.T) {
	g := createTestGame()

	// Pin at X=800, Committee at X=1200 (from createTestGame)
	tests := []struct {
		name     string
		position geometry.Point
		expected bool
	}{
		{"Inside bounds - center", geometry.Point{X: 1000, Y: 2400}, true},
		{"Inside bounds - near pin", geometry.Point{X: 810, Y: 2400}, true},
		{"Inside bounds - near committee", geometry.Point{X: 1190, Y: 2400}, true},
		{"Outside bounds - left of pin", geometry.Point{X: 700, Y: 2400}, false},
		{"Outside bounds - right of committee", geometry.Point{X: 1300, Y: 2400}, false},
		{"At pin exactly", geometry.Point{X: 800, Y: 2400}, true},
		{"At committee exactly", geometry.Point{X: 1200, Y: 2400}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := g.isWithinLineBounds(tt.position)
			if result != tt.expected {
				t.Errorf("isWithinLineBounds(%v) = %v, expected %v",
					tt.position, result, tt.expected)
			}
		})
	}
}
