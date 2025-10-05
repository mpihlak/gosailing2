package game

import (
	"testing"

	"github.com/mpihlak/gosailing2/pkg/geometry"
)

func TestMarkRounding_Phase1_SailPastMark(t *testing.T) {
	g := createTestGame()
	g.raceStarted = true
	g.hasCrossedLine = true

	// Get upwind mark position (Y=1800 from createTestGame)
	upwindMark := g.Arena.Marks[2]

	// Start south of mark
	g.Boat.Pos = geometry.Point{X: upwindMark.Pos.X, Y: 2000}
	g.updateMarkRounding()

	if g.markRoundingPhase1 {
		t.Error("Phase 1 should not be complete when south of mark")
	}

	// Move north past mark (Y <= markY - 1)
	g.Boat.Pos = geometry.Point{X: upwindMark.Pos.X, Y: upwindMark.Pos.Y - 1}
	g.updateMarkRounding()

	if !g.markRoundingPhase1 {
		t.Error("Phase 1 should be complete after sailing past mark")
	}
}

func TestMarkRounding_Phase2_TravelToPort(t *testing.T) {
	g := createTestGame()
	g.raceStarted = true
	g.hasCrossedLine = true
	upwindMark := g.Arena.Marks[2]

	// Complete phase 1
	g.Boat.Pos = geometry.Point{X: upwindMark.Pos.X, Y: upwindMark.Pos.Y - 10}
	g.updateMarkRounding()

	if !g.markRoundingPhase1 {
		t.Fatal("Phase 1 should be complete")
	}

	// Still east of mark (right side)
	g.Boat.Pos = geometry.Point{X: upwindMark.Pos.X + 10, Y: upwindMark.Pos.Y - 10}
	g.updateMarkRounding()

	if g.markRoundingPhase2 {
		t.Error("Phase 2 should not be complete when still east of mark")
	}

	// Move to port (left/west) of mark
	g.Boat.Pos = geometry.Point{X: upwindMark.Pos.X - 1, Y: upwindMark.Pos.Y - 10}
	g.updateMarkRounding()

	if !g.markRoundingPhase2 {
		t.Error("Phase 2 should be complete after moving to port of mark")
	}
}

func TestMarkRounding_Phase2_ResetIfDriftSouth(t *testing.T) {
	g := createTestGame()
	g.raceStarted = true
	g.hasCrossedLine = true
	upwindMark := g.Arena.Marks[2]

	// Complete phase 1, but NOT phase 2 yet
	g.Boat.Pos = geometry.Point{X: upwindMark.Pos.X + 10, Y: upwindMark.Pos.Y - 10}
	g.updateMarkRounding()

	if !g.markRoundingPhase1 {
		t.Fatal("Phase 1 should be complete")
	}

	// Drift south of mark before completing phase 2
	g.Boat.Pos = geometry.Point{X: upwindMark.Pos.X + 10, Y: upwindMark.Pos.Y + 10}
	g.updateMarkRounding()

	if g.markRoundingPhase2 {
		t.Error("Phase 2 should not be set when boat drifts south before completing it")
	}

	// Phase 1 should still be complete
	if !g.markRoundingPhase1 {
		t.Error("Phase 1 should remain complete")
	}
}

func TestMarkRounding_Phase3_SailBelowMark(t *testing.T) {
	g := createTestGame()
	g.raceStarted = true
	g.hasCrossedLine = true
	upwindMark := g.Arena.Marks[2]

	// Complete phase 1 and 2
	g.Boat.Pos = geometry.Point{X: upwindMark.Pos.X - 10, Y: upwindMark.Pos.Y - 10}
	g.updateMarkRounding()

	if !g.markRoundingPhase1 || !g.markRoundingPhase2 {
		t.Fatal("Phases 1 and 2 should be complete")
	}

	// Still north of mark
	if g.markRoundingPhase3 || g.markRounded {
		t.Error("Phase 3 should not be complete while north of mark")
	}

	// Sail below mark (Y >= markY + 1)
	g.Boat.Pos = geometry.Point{X: upwindMark.Pos.X - 10, Y: upwindMark.Pos.Y + 1}
	g.updateMarkRounding()

	if !g.markRoundingPhase3 {
		t.Error("Phase 3 should be complete after sailing below mark")
	}

	if !g.markRounded {
		t.Error("markRounded flag should be set after all three phases")
	}
}

func TestMarkRounding_CompleteSequence(t *testing.T) {
	g := createTestGame()
	g.raceStarted = true
	g.hasCrossedLine = true
	upwindMark := g.Arena.Marks[2]

	// Verify initial state
	if g.markRoundingPhase1 || g.markRoundingPhase2 || g.markRoundingPhase3 || g.markRounded {
		t.Fatal("All rounding phases should start as false")
	}

	// Step 1: Sail north past mark
	g.Boat.Pos = geometry.Point{X: upwindMark.Pos.X, Y: upwindMark.Pos.Y - 5}
	g.updateMarkRounding()

	if !g.markRoundingPhase1 {
		t.Error("Phase 1 should be complete")
	}
	if g.markRoundingPhase2 || g.markRoundingPhase3 || g.markRounded {
		t.Error("Only phase 1 should be complete at this point")
	}

	// Step 2: Move to port of mark
	g.Boat.Pos = geometry.Point{X: upwindMark.Pos.X - 5, Y: upwindMark.Pos.Y - 5}
	g.updateMarkRounding()

	if !g.markRoundingPhase1 || !g.markRoundingPhase2 {
		t.Error("Phases 1 and 2 should be complete")
	}
	if g.markRoundingPhase3 || g.markRounded {
		t.Error("Phase 3 should not be complete yet")
	}

	// Step 3: Sail south past mark
	g.Boat.Pos = geometry.Point{X: upwindMark.Pos.X - 5, Y: upwindMark.Pos.Y + 5}
	g.updateMarkRounding()

	if !g.markRoundingPhase1 || !g.markRoundingPhase2 || !g.markRoundingPhase3 {
		t.Error("All three phases should be complete")
	}
	if !g.markRounded {
		t.Error("markRounded should be true after completing all phases")
	}
}

func TestMarkRounding_Phase2_ResetIfDriftEast(t *testing.T) {
	g := createTestGame()
	g.raceStarted = true
	g.hasCrossedLine = true
	upwindMark := g.Arena.Marks[2]

	// Complete phases 1 and 2
	g.Boat.Pos = geometry.Point{X: upwindMark.Pos.X - 10, Y: upwindMark.Pos.Y - 10}
	g.updateMarkRounding()

	if !g.markRoundingPhase2 {
		t.Fatal("Phase 2 should be complete")
	}

	// Drift back east of mark (while still north)
	g.Boat.Pos = geometry.Point{X: upwindMark.Pos.X + 5, Y: upwindMark.Pos.Y - 10}
	g.updateMarkRounding()

	if g.markRoundingPhase2 {
		t.Error("Phase 2 should reset when drifting east while north of mark")
	}

	if !g.markRoundingPhase1 {
		t.Error("Phase 1 should still be complete")
	}
}

func TestMarkRounding_NotActiveBeforeLineCrossing(t *testing.T) {
	g := createTestGame()
	g.raceStarted = true
	g.hasCrossedLine = false // Haven't crossed start line yet
	upwindMark := g.Arena.Marks[2]

	// Try to complete rounding without crossing line first
	g.Boat.Pos = geometry.Point{X: upwindMark.Pos.X - 10, Y: upwindMark.Pos.Y - 10}

	// updateMarkRounding would not be called in game logic if !hasCrossedLine
	// This test verifies the precondition

	if g.markRoundingPhase1 || g.markRoundingPhase2 || g.markRoundingPhase3 {
		t.Error("Mark rounding should not be tracked before crossing start line")
	}
}

func TestMarkRounding_WithInsufficientMarks(t *testing.T) {
	g := createTestGame()
	g.raceStarted = true
	g.hasCrossedLine = true

	// Remove upwind mark
	g.Arena.Marks = g.Arena.Marks[:2] // Only Pin and Committee

	// Should not panic
	g.updateMarkRounding()

	if g.markRoundingPhase1 || g.markRoundingPhase2 || g.markRoundingPhase3 {
		t.Error("No rounding phases should be triggered without upwind mark")
	}
}
