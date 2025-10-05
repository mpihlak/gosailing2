package dashboard

import (
	"math"
	"testing"
	"time"

	"github.com/mpihlak/gosailing2/pkg/game/objects"
	"github.com/mpihlak/gosailing2/pkg/game/world"
	"github.com/mpihlak/gosailing2/pkg/geometry"
	"github.com/mpihlak/gosailing2/pkg/polars"
)

// Helper to create test dashboard
func createTestDashboard() *Dashboard {
	wind := world.NewOscillatingWind(10.0, 10.0, 2000.0)
	boat := &objects.Boat{
		Pos:     geometry.Point{X: 1000, Y: 2500},
		Heading: 0, // North
		Speed:   6.0,
		Polars:  &polars.RealisticPolar{},
		Wind:    wind,
	}

	return &Dashboard{
		Boat:       boat,
		Wind:       wind,
		StartTime:  time.Now(),
		LineStart:  geometry.Point{X: 800, Y: 2400},
		LineEnd:    geometry.Point{X: 1200, Y: 2400},
		UpwindMark: geometry.Point{X: 1000, Y: 1800},
	}
}

func TestCalculateVMG_Upwind(t *testing.T) {
	dash := createTestDashboard()

	// Boat heading north (0°), wind from north (0°)
	// TWA = 0°, VMG should equal boat speed (perfect upwind)
	dash.Boat.Heading = 0
	dash.Boat.Speed = 6.0

	vmg := dash.CalculateVMG()

	// VMG = Speed * cos(TWA) = 6.0 * cos(0) = 6.0
	if math.Abs(vmg-6.0) > 0.01 {
		t.Errorf("VMG for TWA=0 should be ~6.0, got %.2f", vmg)
	}
}

func TestCalculateVMG_Beat45Degrees(t *testing.T) {
	dash := createTestDashboard()

	// Boat at 45° close-hauled, wind from north (0°)
	// TWA = 45°
	dash.Boat.Heading = 45
	dash.Boat.Speed = 6.0

	vmg := dash.CalculateVMG()

	// VMG = 6.0 * cos(45°) = 6.0 * 0.707... ≈ 4.24
	expected := 6.0 * math.Cos(45*math.Pi/180)

	if math.Abs(vmg-expected) > 0.01 {
		t.Errorf("VMG for TWA=45 should be ~%.2f, got %.2f", expected, vmg)
	}
}

func TestCalculateVMG_BeamReach90Degrees(t *testing.T) {
	dash := createTestDashboard()

	// Boat at 90° (beam reach), wind from north
	// TWA = 90°, VMG towards wind should be 0
	dash.Boat.Heading = 90
	dash.Boat.Speed = 8.0

	vmg := dash.CalculateVMG()

	// VMG = 8.0 * cos(90°) = 8.0 * 0 = 0
	if math.Abs(vmg) > 0.01 {
		t.Errorf("VMG for TWA=90 should be ~0, got %.2f", vmg)
	}
}

func TestCalculateVMG_Downwind180Degrees(t *testing.T) {
	dash := createTestDashboard()

	// Boat heading south (180°), wind from north (0°)
	// TWA = 180° (dead downwind)
	dash.Boat.Heading = 180
	dash.Boat.Speed = 5.0

	vmg := dash.CalculateVMG()

	// VMG = 5.0 * cos(180°) = 5.0 * -1 = -5.0 (negative = away from wind)
	if math.Abs(vmg-(-5.0)) > 0.01 {
		t.Errorf("VMG for TWA=180 should be ~-5.0, got %.2f", vmg)
	}
}

func TestCalculateVMG_NegativeTWA(t *testing.T) {
	dash := createTestDashboard()

	// Boat at -45° (port tack close-hauled), wind from north
	// TWA = -45°
	dash.Boat.Heading = 315 // -45° normalized to 315°
	dash.Boat.Speed = 6.0

	vmg := dash.CalculateVMG()

	// VMG = 6.0 * cos(-45°) = 6.0 * cos(45°) ≈ 4.24
	expected := 6.0 * math.Cos(45*math.Pi/180)

	if math.Abs(vmg-expected) > 0.01 {
		t.Errorf("VMG for TWA=-45 should be ~%.2f, got %.2f", expected, vmg)
	}
}

func TestFindBestVMG_Upwind(t *testing.T) {
	dash := createTestDashboard()

	// Set boat to upwind mode (TWA < 90)
	dash.Boat.Heading = 45 // Upwind

	bestVMG := dash.FindBestVMG()

	// Best upwind VMG should be positive
	if bestVMG <= 0 {
		t.Errorf("Best upwind VMG should be positive, got %.2f", bestVMG)
	}

	// Should be achievable (within reasonable range for 10 kts wind)
	if bestVMG > 10.0 {
		t.Errorf("Best VMG seems unreasonably high: %.2f", bestVMG)
	}
}

func TestFindBestVMG_Downwind(t *testing.T) {
	dash := createTestDashboard()

	// Set boat to downwind mode (TWA > 90)
	dash.Boat.Heading = 135 // Broad reach

	bestVMG := dash.FindBestVMG()

	// Best downwind VMG should be negative (away from wind)
	if bestVMG >= 0 {
		t.Errorf("Best downwind VMG should be negative, got %.2f", bestVMG)
	}
}

func TestCalculateDistanceToLine_BelowLine(t *testing.T) {
	dash := createTestDashboard()

	// Boat below line (Y > lineY)
	dash.Boat.Pos = geometry.Point{X: 1000, Y: 2500}

	distance := dash.CalculateDistanceToLine()

	// Should be positive when below line (pre-start side)
	if distance <= 0 {
		t.Errorf("Distance should be positive when below line, got %.2f", distance)
	}

	// Should be approximately 100 meters (account for bow offset)
	if math.Abs(distance-100) > 10.0 {
		t.Errorf("Distance should be ~100m, got %.2f", distance)
	}
}

func TestCalculateDistanceToLine_AboveLine(t *testing.T) {
	dash := createTestDashboard()

	// Boat above line (Y < lineY) - course side
	dash.Boat.Pos = geometry.Point{X: 1000, Y: 2300}

	distance := dash.CalculateDistanceToLine()

	// Should be negative when above line (course side)
	if distance >= 0 {
		t.Errorf("Distance should be negative when above line, got %.2f", distance)
	}

	// Should be approximately -100 meters (account for bow offset)
	if math.Abs(distance-(-100)) > 10.0 {
		t.Errorf("Distance should be ~-100m, got %.2f", distance)
	}
}

func TestCalculateDistanceToLine_OnLine(t *testing.T) {
	dash := createTestDashboard()

	// Boat exactly on line
	dash.Boat.Pos = geometry.Point{X: 1000, Y: 2400}

	distance := dash.CalculateDistanceToLine()

	// Should be very close to 0
	if math.Abs(distance) > 10.0 {
		t.Errorf("Distance should be ~0 when on line, got %.2f", distance)
	}
}

func TestCalculateVMG_ZeroSpeed(t *testing.T) {
	dash := createTestDashboard()

	// Boat not moving
	dash.Boat.Speed = 0.0
	dash.Boat.Heading = 45

	vmg := dash.CalculateVMG()

	if vmg != 0 {
		t.Errorf("VMG should be 0 when speed is 0, got %.2f", vmg)
	}
}

func TestCalculateVMG_NoNaN(t *testing.T) {
	dash := createTestDashboard()

	// Test various headings to ensure no NaN
	headings := []float64{0, 45, 90, 135, 180, 225, 270, 315}

	for _, heading := range headings {
		dash.Boat.Heading = heading
		vmg := dash.CalculateVMG()

		if math.IsNaN(vmg) || math.IsInf(vmg, 0) {
			t.Errorf("VMG should not be NaN/Inf for heading %.0f, got %.2f", heading, vmg)
		}
	}
}

func TestFindBestVMG_Consistency(t *testing.T) {
	dash := createTestDashboard()

	// Upwind: best VMG should be same regardless of which tack
	dash.Boat.Heading = 45
	bestVMG1 := dash.FindBestVMG()

	dash.Boat.Heading = 315 // -45°
	bestVMG2 := dash.FindBestVMG()

	// Should be very close (may not be exact due to polar curve)
	if math.Abs(bestVMG1-bestVMG2) > 0.1 {
		t.Errorf("Best VMG should be similar for starboard (%.2f) and port (%.2f) tack",
			bestVMG1, bestVMG2)
	}
}
