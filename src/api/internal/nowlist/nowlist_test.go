package nowlist

import (
	"testing"
)

func TestFilterNowGoals(t *testing.T) {
	goals := []Goal{
		{ID: 1, Now: true},
		{ID: 2, Now: false},
		{ID: 3, Now: true},
	}

	filtered := FilterNowGoals(goals)
	if len(filtered) != 2 {
		t.Fatalf("ville ha 2 mål, fick %d", len(filtered))
	}
	if filtered[0].ID != 1 || filtered[1].ID != 3 {
		t.Fatalf("fel ordning: %v", filtered)
	}
}

func TestFilterByPlace(t *testing.T) {
	goals := []Goal{
		{ID: 1, PlaceIDs: []int64{10}},
		{ID: 2, PlaceIDs: []int64{20}},
		{ID: 3, PlaceIDs: []int64{10, 20}},
		{ID: 4},
	}

	place := int64(10)
	filtered := FilterByPlace(goals, &place)
	if len(filtered) != 3 {
		t.Fatalf("ville ha 3 mål, fick %d", len(filtered))
	}
	for _, goal := range filtered {
		if goal.ID == 2 {
			t.Fatal("mål 2 gäller inte den valda platsen")
		}
	}

	// Utan vald plats filtreras ingenting bort.
	if got := len(FilterByPlace(goals, nil)); got != 4 {
		t.Fatalf("ville ha 4 mål utan vald plats, fick %d", got)
	}
}

func TestBuildOrdersChildrenUnderParents(t *testing.T) {
	goals := []Goal{
		{ID: 1, Now: true},
		{ID: 2, Now: true, ParentIDs: []int64{1}},
		{ID: 3, Now: true},
	}

	rows := Build(goals, nil)
	if len(rows) != 3 {
		t.Fatalf("ville ha 3 rader, fick %d", len(rows))
	}

	if rows[0].Goal.ID != 1 || rows[0].Depth != 0 || rows[0].Number != 1 {
		t.Fatalf("rad 0 fel: %+v", rows[0])
	}
	if rows[1].Goal.ID != 2 || rows[1].Depth != 1 {
		t.Fatalf("rad 1 fel: %+v", rows[1])
	}
	if rows[1].ParentID == nil || *rows[1].ParentID != 1 {
		t.Fatalf("rad 1 saknar förälder: %+v", rows[1])
	}
	if rows[1].Number != 0 {
		t.Fatalf("barnmål ska inte ha nummer: %+v", rows[1])
	}
	if rows[2].Goal.ID != 3 || rows[2].Number != 2 {
		t.Fatalf("rad 2 fel: %+v", rows[2])
	}
}

func TestBuildHidesFilteredParent(t *testing.T) {
	// Föräldern är inte markerad med nu-märket och syns därför inte.
	// Barnet ska då ligga på toppnivå.
	goals := []Goal{
		{ID: 1, Now: false},
		{ID: 2, Now: true, ParentIDs: []int64{1}},
	}

	rows := Build(goals, nil)
	if len(rows) != 1 {
		t.Fatalf("ville ha 1 rad, fick %d", len(rows))
	}
	if rows[0].Depth != 0 || rows[0].ParentID != nil {
		t.Fatalf("barnet ska ligga på toppnivå: %+v", rows[0])
	}
}

func TestBuildUsesNearestVisibleAncestor(t *testing.T) {
	// 1 -> 2 -> 3. Bara 1 och 3 syns, så 3 hänger under 1.
	goals := []Goal{
		{ID: 1, Now: true},
		{ID: 2, Now: false, ParentIDs: []int64{1}},
		{ID: 3, Now: true, ParentIDs: []int64{2}},
	}

	rows := Build(goals, nil)
	if len(rows) != 2 {
		t.Fatalf("ville ha 2 rader, fick %d", len(rows))
	}
	if rows[1].Goal.ID != 3 || rows[1].Depth != 1 {
		t.Fatalf("rad 1 fel: %+v", rows[1])
	}
	if rows[1].ParentID == nil || *rows[1].ParentID != 1 {
		t.Fatalf("rad 1 ska hänga under 1: %+v", rows[1])
	}
}

func TestBuildBreaksCycles(t *testing.T) {
	// 1 -> 2 -> 1. Cykeln bryts så att målen inte försvinner.
	goals := []Goal{
		{ID: 1, Now: true, ParentIDs: []int64{2}},
		{ID: 2, Now: true, ParentIDs: []int64{1}},
	}

	rows := Build(goals, nil)
	if len(rows) != 2 {
		t.Fatalf("ville ha 2 rader, fick %d", len(rows))
	}
}

func TestBuildAppliesPlaceFilter(t *testing.T) {
	goals := []Goal{
		{ID: 1, Now: true, PlaceIDs: []int64{10}},
		{ID: 2, Now: true, PlaceIDs: []int64{20}},
		{ID: 3, Now: true},
	}

	place := int64(10)
	rows := Build(goals, &place)
	if len(rows) != 2 {
		t.Fatalf("ville ha 2 rader, fick %d", len(rows))
	}
	for _, row := range rows {
		if row.Goal.ID == 2 {
			t.Fatal("mål 2 gäller inte den valda platsen")
		}
	}
}
