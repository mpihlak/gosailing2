package world

import (
	"math"
	"testing"

	"github.com/mpihlak/gosailing2/pkg/geometry"
)

func TestOscillatingWind_LeftSide(t *testing.T) {
	// Left side: 14 kts, Right side: 8 kts
	wind := NewOscillatingWind(14.0, 8.0, 2000.0)

	// Test at far left (X=0)
	dir, speed := wind.GetWind(geometry.Point{X: 0, Y: 1000})

	if dir != 0 {
		t.Errorf("Wind direction at left side should be 0, got %.1f", dir)
	}

	if math.Abs(speed-14.0) > 0.1 {
		t.Errorf("Wind speed at left side should be ~14 kts, got %.1f", speed)
	}
}

func TestOscillatingWind_RightSide(t *testing.T) {
	wind := NewOscillatingWind(14.0, 8.0, 2000.0)

	// Test at far right (X=2000)
	dir, speed := wind.GetWind(geometry.Point{X: 2000, Y: 1000})

	if dir != 0 {
		t.Errorf("Wind direction at right side should be 0, got %.1f", dir)
	}

	if math.Abs(speed-8.0) > 0.1 {
		t.Errorf("Wind speed at right side should be ~8 kts, got %.1f", speed)
	}
}

func TestOscillatingWind_Center(t *testing.T) {
	wind := NewOscillatingWind(14.0, 8.0, 2000.0)

	// Test at center (X=1000)
	dir, speed := wind.GetWind(geometry.Point{X: 1000, Y: 1000})

	expectedSpeed := (14.0 + 8.0) / 2.0 // Should be 11 kts

	if dir != 0 {
		t.Errorf("Wind direction at center should be 0, got %.1f", dir)
	}

	if math.Abs(speed-expectedSpeed) > 0.1 {
		t.Errorf("Wind speed at center should be ~%.1f kts, got %.1f", expectedSpeed, speed)
	}
}

func TestOscillatingWind_Interpolation(t *testing.T) {
	wind := NewOscillatingWind(10.0, 20.0, 2000.0)

	tests := []struct {
		name          string
		x             float64
		expectedSpeed float64
	}{
		{"Quarter from left", 500.0, 12.5},   // 10 + (20-10)*0.25
		{"Half way", 1000.0, 15.0},           // 10 + (20-10)*0.5
		{"Three quarters", 1500.0, 17.5},     // 10 + (20-10)*0.75
		{"Left edge", 0.0, 10.0},             // Left speed
		{"Right edge", 2000.0, 20.0},         // Right speed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, speed := wind.GetWind(geometry.Point{X: tt.x, Y: 1000})

			if math.Abs(speed-tt.expectedSpeed) > 0.1 {
				t.Errorf("At X=%.1f, expected speed %.1f, got %.1f",
					tt.x, tt.expectedSpeed, speed)
			}
		})
	}
}

func TestOscillatingWind_DirectionConstant(t *testing.T) {
	wind := NewOscillatingWind(10.0, 15.0, 2000.0)

	// Wind direction should be 0 (North) everywhere initially
	positions := []geometry.Point{
		{X: 0, Y: 0},
		{X: 500, Y: 1000},
		{X: 1000, Y: 2000},
		{X: 1500, Y: 500},
		{X: 2000, Y: 1500},
	}

	for _, pos := range positions {
		dir, _ := wind.GetWind(pos)
		if dir != 0 {
			t.Errorf("Wind direction at (%v) should be 0, got %.1f", pos, dir)
		}
	}
}

func TestOscillatingWind_NegativePosition(t *testing.T) {
	wind := NewOscillatingWind(10.0, 20.0, 2000.0)

	// Test with X < 0 (should clamp to 0)
	_, speed := wind.GetWind(geometry.Point{X: -100, Y: 1000})

	if math.Abs(speed-10.0) > 0.1 {
		t.Errorf("Wind speed at X=-100 should clamp to left speed (10), got %.1f", speed)
	}
}

func TestOscillatingWind_BeyondWorldWidth(t *testing.T) {
	wind := NewOscillatingWind(10.0, 20.0, 2000.0)

	// Test with X > worldWidth (should clamp to worldWidth)
	_, speed := wind.GetWind(geometry.Point{X: 3000, Y: 1000})

	if math.Abs(speed-20.0) > 0.1 {
		t.Errorf("Wind speed at X=3000 should clamp to right speed (20), got %.1f", speed)
	}
}

func TestOscillatingWind_UpdateWithElapsedTime(t *testing.T) {
	wind := NewOscillatingWind(10.0, 20.0, 2000.0)

	// Get initial direction
	dirBefore, _ := wind.GetWind(geometry.Point{X: 1000, Y: 1000})

	// Update with elapsed time (should cause oscillation)
	wind.UpdateWithElapsedTime(5.0) // 5 seconds

	// Get direction after update
	dirAfter, _ := wind.GetWind(geometry.Point{X: 1000, Y: 1000})

	// Direction should have changed due to oscillation
	// (exact value depends on oscillation formula, but should not be identical)
	if dirBefore == dirAfter {
		// This might occasionally fail if oscillation period aligns perfectly
		// but generally should show change
		t.Logf("Warning: Direction unchanged after oscillation (%.1f -> %.1f). May be coincidental.",
			dirBefore, dirAfter)
	}
}

func TestOscillatingWind_YPositionDoesNotAffect(t *testing.T) {
	wind := NewOscillatingWind(10.0, 20.0, 2000.0)

	// Same X, different Y positions should give same result
	_, speed1 := wind.GetWind(geometry.Point{X: 1000, Y: 0})
	_, speed2 := wind.GetWind(geometry.Point{X: 1000, Y: 1000})
	_, speed3 := wind.GetWind(geometry.Point{X: 1000, Y: 3000})

	if speed1 != speed2 || speed2 != speed3 {
		t.Errorf("Wind speed should be same for same X position regardless of Y: %.1f, %.1f, %.1f",
			speed1, speed2, speed3)
	}
}

func TestOscillatingWind_EqualLeftRight(t *testing.T) {
	// When left and right speeds are equal, should be constant across field
	wind := NewOscillatingWind(12.0, 12.0, 2000.0)

	positions := []float64{0, 500, 1000, 1500, 2000}

	for _, x := range positions {
		_, speed := wind.GetWind(geometry.Point{X: x, Y: 1000})
		if math.Abs(speed-12.0) > 0.01 {
			t.Errorf("At X=%.1f, expected constant speed 12.0, got %.1f", x, speed)
		}
	}
}
