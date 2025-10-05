package world

import (
	"testing"

	"github.com/mpihlak/gosailing2/pkg/geometry"
)

func TestCheckCollisions_DirectHit(t *testing.T) {
	arena := &Arena{
		Marks: []*Mark{
			{Pos: geometry.Point{X: 1000, Y: 2400}, Name: "Pin"},
			{Pos: geometry.Point{X: 1200, Y: 2400}, Name: "Committee"},
		},
	}

	// Boat at same position as Pin mark - should collide
	collisions := arena.CheckCollisions(
		geometry.Point{X: 1000, Y: 2400},
		5.0, // boat radius
	)

	if len(collisions) != 1 {
		t.Fatalf("Expected 1 collision, got %d", len(collisions))
	}

	if collisions[0].MarkName != "Pin" {
		t.Errorf("Expected collision with Pin, got %s", collisions[0].MarkName)
	}

	if collisions[0].Type != CollisionMark {
		t.Errorf("Expected CollisionMark type, got %v", collisions[0].Type)
	}
}

func TestCheckCollisions_NearMiss(t *testing.T) {
	arena := &Arena{
		Marks: []*Mark{
			{Pos: geometry.Point{X: 1000, Y: 2400}, Name: "Pin"},
		},
	}

	// Boat 10m away from mark - should not collide
	// (boat radius 5.0 + mark radius 0.5 = 5.5m threshold)
	collisions := arena.CheckCollisions(
		geometry.Point{X: 1010, Y: 2400},
		5.0,
	)

	if len(collisions) != 0 {
		t.Errorf("Expected no collision, got %d collisions", len(collisions))
	}
}

func TestCheckCollisions_EdgeCase(t *testing.T) {
	arena := &Arena{
		Marks: []*Mark{
			{Pos: geometry.Point{X: 1000, Y: 2400}, Name: "Pin"},
		},
	}

	// Boat exactly at collision threshold distance
	// distance = boatRadius + markRadius = 5.0 + 0.5 = 5.5m
	collisions := arena.CheckCollisions(
		geometry.Point{X: 1005.5, Y: 2400},
		5.0,
	)

	// Should NOT collide (< threshold, not <=)
	if len(collisions) != 0 {
		t.Errorf("Expected no collision at threshold, got %d", len(collisions))
	}

	// Just inside threshold - should collide
	collisions = arena.CheckCollisions(
		geometry.Point{X: 1005.4, Y: 2400},
		5.0,
	)

	if len(collisions) != 1 {
		t.Errorf("Expected collision just inside threshold, got %d", len(collisions))
	}
}

func TestCheckCollisions_MultipleMarks(t *testing.T) {
	arena := &Arena{
		Marks: []*Mark{
			{Pos: geometry.Point{X: 1000, Y: 2400}, Name: "Pin"},
			{Pos: geometry.Point{X: 1200, Y: 2400}, Name: "Committee"},
			{Pos: geometry.Point{X: 1100, Y: 1800}, Name: "Upwind"},
		},
	}

	// Boat near Pin only
	collisions := arena.CheckCollisions(
		geometry.Point{X: 1002, Y: 2400},
		5.0,
	)

	if len(collisions) != 1 {
		t.Fatalf("Expected 1 collision, got %d", len(collisions))
	}

	if collisions[0].MarkName != "Pin" {
		t.Errorf("Expected Pin collision, got %s", collisions[0].MarkName)
	}
}

func TestCheckCollisions_NoMarks(t *testing.T) {
	arena := &Arena{
		Marks: []*Mark{},
	}

	collisions := arena.CheckCollisions(
		geometry.Point{X: 1000, Y: 2400},
		5.0,
	)

	if len(collisions) != 0 {
		t.Errorf("Expected no collisions with empty arena, got %d", len(collisions))
	}
}

func TestCheckCollisions_DiagonalDistance(t *testing.T) {
	arena := &Arena{
		Marks: []*Mark{
			{Pos: geometry.Point{X: 1000, Y: 2400}, Name: "Pin"},
		},
	}

	// Boat at diagonal distance: sqrt(3^2 + 4^2) = 5m from mark
	// Should collide since 5.0 < (5.0 + 0.5)
	collisions := arena.CheckCollisions(
		geometry.Point{X: 1003, Y: 2404},
		5.0,
	)

	if len(collisions) != 1 {
		t.Errorf("Expected collision at diagonal distance, got %d", len(collisions))
	}
}
