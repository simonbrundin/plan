package handlers

import (
	"context"
	"net/http"
	"sort"
	"sync"

	"plan-api/internal/models"

	"github.com/gin-gonic/gin"
)

// goalGraph är en användares hela målgraph: målen, relationerna mellan
// dem och beroendena mellan dem.
//
// Den byggs med ett fåtal frågor i stället för en fråga per mål. Mot Neon
// kostar varje rundtur ungefär 40 ms, så en målsida med 28 undermål gick
// från 36 rundturer (cirka 2 sekunder) till fyra (cirka 0,2 sekunder).
type goalGraph struct {
	goals        []models.Goal
	relations    []models.GoalRelation
	dependencies []models.GoalDependency

	goalsByID map[int64]models.Goal
	// childrenOf map parent -> relationer, sorterade på order.
	childrenOf map[int64][]models.GoalRelation
	// parentsOf map child -> relationer.
	parentsOf map[int64][]models.GoalRelation
	// dependsOn map goal -> beroendets mål.
	dependsOn map[int64][]models.Goal
	// blocking map goal -> mål som beror på det.
	blocking map[int64][]models.Goal
}

// loadGraph hämtar hela grafen för en användare.
//
// Frågorna körs parallellt - de är oberoende av varandra, och Neon har
// en rundtur per fråga. Platser hämtas i en extra fråga och fylls i på
// målen.
func (h *GoalHandler) loadGraph(ctx context.Context, userID int64) (*goalGraph, error) {
	var (
		goals        []models.Goal
		relations    []models.GoalRelation
		dependencies []models.GoalDependency
		goalsErr     error
		relationsErr error
		depsErr      error
		wg           sync.WaitGroup
	)

	wg.Add(3)
	go func() {
		defer wg.Done()
		goals, goalsErr = h.repo.GetAllGoals(ctx, userID)
	}()
	go func() {
		defer wg.Done()
		relations, relationsErr = h.repo.GetAllRelations(ctx, userID)
	}()
	go func() {
		defer wg.Done()
		dependencies, depsErr = h.repo.GetAllDependencies(ctx, userID)
	}()
	wg.Wait()

	if goalsErr != nil {
		return nil, goalsErr
	}
	if relationsErr != nil {
		return nil, relationsErr
	}
	if depsErr != nil {
		return nil, depsErr
	}

	if goals == nil {
		goals = []models.Goal{}
	}
	if relations == nil {
		relations = []models.GoalRelation{}
	}
	if dependencies == nil {
		dependencies = []models.GoalDependency{}
	}

	// Platser hämtas samlat för alla mål - en fråga i stället för en per mål.
	h.attachPlaces(ctx, goals)

	graph := &goalGraph{
		goals:        goals,
		relations:    relations,
		dependencies: dependencies,
		goalsByID:    make(map[int64]models.Goal, len(goals)),
		childrenOf:   map[int64][]models.GoalRelation{},
		parentsOf:    map[int64][]models.GoalRelation{},
		dependsOn:    map[int64][]models.Goal{},
		blocking:     map[int64][]models.Goal{},
	}

	for _, g := range goals {
		graph.goalsByID[g.ID] = g
	}

	for _, rel := range relations {
		graph.childrenOf[rel.ParentID] = append(graph.childrenOf[rel.ParentID], rel)
		graph.parentsOf[rel.ChildID] = append(graph.parentsOf[rel.ChildID], rel)
	}

	// Barnen sorteras på "order" här, så att ordningen inte hänger på hur
	// databasen råkar returnera raderna. Barn med samma ordning sorteras på
	// id, så att resultatet blir detsamma varje gång.
	for parentID, rels := range graph.childrenOf {
		sort.SliceStable(rels, func(i, j int) bool {
			if rels[i].Order != rels[j].Order {
				return rels[i].Order < rels[j].Order
			}
			return rels[i].ChildID < rels[j].ChildID
		})
		graph.childrenOf[parentID] = rels
	}

	for _, dep := range dependencies {
		dependsOn, ok := graph.goalsByID[dep.DependsOnID]
		if !ok {
			continue
		}
		goal, ok := graph.goalsByID[dep.GoalID]
		if !ok {
			continue
		}
		graph.dependsOn[dep.GoalID] = append(graph.dependsOn[dep.GoalID], dependsOn)
		graph.blocking[dep.DependsOnID] = append(graph.blocking[dep.DependsOnID], goal)
	}

	return graph, nil
}

// goal returnerar ett mål ur grafen. Ok är false om målet inte finns
// bland användarens mål.
func (g *goalGraph) goal(goalID int64) (models.Goal, bool) {
	goal, ok := g.goalsByID[goalID]
	return goal, ok
}

// childRelations returnerar relationerna till ett måls barn, sorterade på
// order.
func (g *goalGraph) childRelations(goalID int64) []models.ChildRelation {
	relations := g.childrenOf[goalID]
	out := make([]models.ChildRelation, 0, len(relations))
	for _, rel := range relations {
		out = append(out, models.ChildRelation{
			ChildID: rel.ChildID,
			Order:   rel.Order,
			Weight:  rel.Weight,
		})
	}
	return out
}

// children returnerar ett måls barn med ordning, vikt och beroenden.
func (g *goalGraph) children(goalID int64) []models.ChildWithDeps {
	relations := g.childrenOf[goalID]
	children := make([]models.ChildWithDeps, 0, len(relations))

	for _, rel := range relations {
		child, ok := g.goalsByID[rel.ChildID]
		if !ok {
			continue
		}
		children = append(children, models.ChildWithDeps{
			Goal:      child,
			Order:     rel.Order,
			Weight:    rel.Weight,
			DependsOn: g.dependsOnOf(rel.ChildID),
			Blocking:  g.blockingOf(rel.ChildID),
		})
	}

	return children
}

// parents returnerar ett måls föräldrar.
func (g *goalGraph) parents(goalID int64) []models.Goal {
	relations := g.parentsOf[goalID]
	parents := make([]models.Goal, 0, len(relations))

	for _, rel := range relations {
		parent, ok := g.goalsByID[rel.ParentID]
		if !ok {
			continue
		}
		parents = append(parents, parent)
	}

	return parents
}

// dependenciesOf returnerar de beroenderaden som gäller ett mål.
func (g *goalGraph) dependenciesOf(goalID int64) []models.GoalDependency {
	deps := make([]models.GoalDependency, 0)
	for _, dep := range g.dependencies {
		if dep.GoalID == goalID {
			deps = append(deps, dep)
		}
	}
	return deps
}

// dependsOnOf returnerar de mål som ett mål beror på.
//
// Listan är aldrig nil, så att svaret alltid innehåller [] i stället för
// null - samma sätt som platser och mål.
func (g *goalGraph) dependsOnOf(goalID int64) []models.Goal {
	deps := g.dependsOn[goalID]
	if deps == nil {
		return []models.Goal{}
	}
	return deps
}

// blockingOf returnerar de mål som beror på ett mål. Listan är aldrig nil.
func (g *goalGraph) blockingOf(goalID int64) []models.Goal {
	blocking := g.blocking[goalID]
	if blocking == nil {
		return []models.Goal{}
	}
	return blocking
}

// GetGoalGraph godoc
// @Summary Get the user's whole goal graph
// @Description Returns all goals, relations and dependencies in one response.
// @Tags goals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.GoalGraphResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/goals/graph [get]
func (h *GoalHandler) GetGoalGraph(c *gin.Context) {
	if !h.checkDb(c) {
		return
	}

	userID := getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Message: "Unauthorized"})
		return
	}

	graph, err := h.loadGraph(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.GoalGraphResponse{
		Goals:        graph.goals,
		Relations:    graph.relations,
		Dependencies: graph.dependencies,
	})
}
