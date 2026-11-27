// Package nowlist bygger "Gör nu"-listan.
//
// Logiken här är spegelbilden av gränssnittets lista
// (src/nuxt/app/utils/nowHierarchy.ts och src/nuxt/app/utils/goalFilters.ts).
// API:t och gränssnittet ska visa samma sak, så reglerna ligger här och
// används av båda sidor: gränssnittet via /api/v1/goals/now och externa
// klienter (t.ex. skrivbordswidgeten) via samma endpoint.
package nowlist

import "time"

// noParentID används som nyckel i mappar där "ingen förälder" ska hållas
// isär. Målens id är alltid positiva, så 0 kan inte kollidera.
const noParentID int64 = 0

// Goal är den del av ett mål som behövs för att bygga listan.
type Goal struct {
	ID int64
	// Now markerar att målet är markerat med nu-märket och ska visas
	// på "Gör nu".
	Now bool
	// Started är tidsstämpeln för när arbetet började. Null = inte påbörjat.
	Started *time.Time
	// ParentIDs är målets föräldrar, sorterade på vikt (tyngsta först).
	ParentIDs []int64
	// PlaceIDs är de platser målet gäller. Tomt = målet gäller överallt.
	PlaceIDs []int64
}

// Row är en rad i listan.
type Row struct {
	// Goal är målet på raden.
	Goal Goal
	// Depth är hur djupt raden ligger i hierarkin. 0 = toppnivå.
	Depth int
	// ParentID är föräldern raden visas under. Null = toppnivå.
	ParentID *int64
	// Number är radens nummer bland toppnivåmålen. 0 = raden är ett barnmål.
	Number int
}

// FilterNowGoals tar de mål som är markerade med nu-märket.
//
// Nu-märket styr vilka mål som ligger i "Gör nu" - mål som bara är
// påbörjade visas inte automatiskt.
func FilterNowGoals(goals []Goal) []Goal {
	filtered := make([]Goal, 0, len(goals))
	for _, goal := range goals {
		if goal.Now {
			filtered = append(filtered, goal)
		}
	}
	return filtered
}

// FilterByPlace tar bort mål som inte går att göra på den valda platsen.
//
// Flera platser på ett mål har OR-semantik - det räcker att vara på
// någon av dem. Ett mål utan platser är inte bundet till någon plats och
// visas därför alltid.
//
// Om ingen plats är vald (placeID är nil) filtreras ingenting bort. Det
// är annars en platslös vy - "görbar här" ska inte tömma listan bara
// för att användaren inte har valt någon plats.
func FilterByPlace(goals []Goal, placeID *int64) []Goal {
	if placeID == nil {
		return goals
	}

	filtered := make([]Goal, 0, len(goals))
	for _, goal := range goals {
		if len(goal.PlaceIDs) == 0 {
			filtered = append(filtered, goal)
			continue
		}

		atPlace := false
		for _, id := range goal.PlaceIDs {
			if id == *placeID {
				atPlace = true
				break
			}
		}
		if atPlace {
			filtered = append(filtered, goal)
		}
	}
	return filtered
}

// Build bygger listan som visas på "Gör nu".
//
// Målen kommer in i den ordning de ska sorteras i. Resultatet är samma
// ordning: toppnivåmål i listordning med sina barn indenterade under sig.
func Build(goals []Goal, placeID *int64) []Row {
	parentOf := buildHierarchy(goals)

	visible := FilterByPlace(FilterNowGoals(goals), placeID)
	visibleIDs := make(map[int64]bool, len(visible))
	for _, goal := range visible {
		visibleIDs[goal.ID] = true
	}

	// Varje synligt mål hänger under sin närmaste synliga förälder. Finns
	// ingen sådan förälder ligger målet på toppnivå.
	childrenOf := make(map[int64][]int64)
	for _, goal := range visible {
		parent := nearestVisibleAncestor(goal.ID, parentOf, visibleIDs)
		key := noParentID
		if parent != nil {
			key = *parent
		}
		childrenOf[key] = append(childrenOf[key], goal.ID)
	}

	byID := make(map[int64]Goal, len(visible))
	for _, goal := range visible {
		byID[goal.ID] = goal
	}

	var rows []Row
	number := 0

	// Djup-först: toppnivåmålen först, sedan deras barn, och så vidare.
	var visit func(parentID int64, depth int)
	visit = func(parentID int64, depth int) {
		for _, id := range childrenOf[parentID] {
			goal := byID[id]

			var displayParent *int64
			if parentID != noParentID {
				parent := parentID
				displayParent = &parent
			}

			rowNumber := 0
			if depth == 0 {
				number++
				rowNumber = number
			}

			rows = append(rows, Row{
				Goal:     goal,
				Depth:    depth,
				ParentID: displayParent,
				Number:   rowNumber,
			})

			visit(id, depth+1)
		}
	}
	visit(noParentID, 0)

	return rows
}

// buildHierarchy räknar ut vilken förälder varje mål hänger under.
//
// En förälder som inte finns i listan ignoreras - då räknas målet som
// ett toppnivålsmål. Cykler bryts, så att listan aldrig tappar mål.
func buildHierarchy(goals []Goal) map[int64]int64 {
	ids := make(map[int64]bool, len(goals))
	for _, goal := range goals {
		ids[goal.ID] = true
	}

	parentOf := make(map[int64]int64, len(goals))
	for _, goal := range goals {
		parent := noParentID
		for _, candidate := range goal.ParentIDs {
			if ids[candidate] {
				parent = candidate
				break
			}
		}
		parentOf[goal.ID] = parent
	}

	breakCycles(goals, parentOf)
	return parentOf
}

// breakCycles bryter cykler i förälderekrelationerna.
//
// Om A är förälder till B och B är förälder till A finns ingen toppnivå
// att hänga upp trädet på, och målen skulle försvinna ur listan. Den
// kedjan bryts då genom att det siste målet i kedjan blir toppnivå.
func breakCycles(goals []Goal, parentOf map[int64]int64) {
	for _, start := range goals {
		seen := map[int64]bool{start.ID: true}
		current := parentOf[start.ID]

		for current != noParentID {
			if seen[current] {
				parentOf[current] = noParentID
				break
			}
			seen[current] = true
			current = parentOf[current]
		}
	}
}

// nearestVisibleAncestor ger den närmaste föräldern som också syns i listan.
//
// Om ingen förälder syns blir svaret 0 - då ligger målet på toppnivå.
func nearestVisibleAncestor(id int64, parentOf map[int64]int64, visible map[int64]bool) *int64 {
	seen := map[int64]bool{id: true}
	current := parentOf[id]

	for current != noParentID && !visible[current] {
		if seen[current] {
			return nil
		}
		seen[current] = true
		current = parentOf[current]
	}

	if current == noParentID {
		return nil
	}
	return &current
}
