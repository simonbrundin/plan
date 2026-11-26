package goaldefinition

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"plan/agent/internal/definer"
	"plan/agent/internal/planclient"
)

// TestBuildFyllerLuckor kontrollerar att agentens förslag fyller de
// fält som saknas - och bara dem.
func TestBuildFyllerLuckor(t *testing.T) {
	goal := planclient.Goal{
		ID:    42,
		Title: "Gör något viktigt",
		Tags:  []string{"ute"},
	}

	proposal := Build(Context{
		Goal: goal,
		Definition: definer.Definition{
			Title:            "Ett definierat mål",
			Description:      "Tydliga krav och kriterier.",
			Icon:             "heroicons:flag",
			Due:              "2026-12-01",
			EstimatedMinutes: 120,
			Tags:             []string{"helg"},
		},
		NextPhase: "plan",
	})

	if proposal.Values.Title == nil || *proposal.Values.Title != "Ett definierat mål" {
		t.Errorf("titel = %v", proposal.Values.Title)
	}
	if proposal.Values.Icon == nil || *proposal.Values.Icon != "heroicons:flag" {
		t.Errorf("ikon = %v", proposal.Values.Icon)
	}
	if proposal.Values.Due == nil || *proposal.Values.Due != "2026-12-01" {
		t.Errorf("due = %v", proposal.Values.Due)
	}
	if proposal.Values.EstimatedMinutes == nil || *proposal.Values.EstimatedMinutes != 120 {
		t.Errorf("uppskattning = %v", proposal.Values.EstimatedMinutes)
	}

	// Taggarna slås ihop - målets egen tagg får inte försvinna.
	if len(proposal.Values.Tags) != 2 {
		t.Fatalf("taggar = %v, vill ha både ute och helg", proposal.Values.Tags)
	}
	if proposal.Values.Tags[0] != "ute" || proposal.Values.Tags[1] != "helg" {
		t.Errorf("taggar = %v", proposal.Values.Tags)
	}

	if !proposal.HasDataFields() {
		t.Error("förslaget har datafält")
	}
}

// TestBuildLämnarBefintligaVärden kontrollerar att agenten inte
// skriver över värden som användaren redan angett.
func TestBuildLämnarBefintligaVärden(t *testing.T) {
	due := "2026-11-01"
	icon := "heroicons:star"
	minutes := 30
	rule := "FREQ=DAILY"

	goal := planclient.Goal{
		ID:               42,
		Title:            "Ett mål",
		Icon:             icon,
		Due:              &due,
		EstimatedMinutes: &minutes,
		RecurrenceRule:   &rule,
		Now:              true,
		Eternal:          true,
		Agent:            true,
	}

	proposal := Build(Context{
		Goal: goal,
		Definition: definer.Definition{
			Title:            "Ett mål",
			Description:      "Beskrivning",
			Icon:             "heroicons:flag",
			Due:              "2026-12-24",
			StartDate:        "2026-12-01",
			EstimatedMinutes: 999,
			RecurrenceRule:   "FREQ=WEEKLY;BYDAY=MO",
			Now:              false,
			Eternal:          false,
			Agent:            false,
		},
		NextPhase: "plan",
	})

	if proposal.Values.Due != nil {
		t.Errorf("due = %v, vill ha oförändrat", *proposal.Values.Due)
	}
	if proposal.Values.Icon != nil {
		t.Errorf("ikon = %v, vill ha oförändrad", *proposal.Values.Icon)
	}
	if proposal.Values.EstimatedMinutes != nil {
		t.Errorf("uppskattning = %v, vill ha oförändrad", *proposal.Values.EstimatedMinutes)
	}
	if proposal.Values.RecurrenceRule != nil {
		t.Errorf("upprepningsregel = %v, vill ha oförändrad", *proposal.Values.RecurrenceRule)
	}
	if proposal.Values.Now != nil {
		t.Errorf("now = %v, vill ha oförändrad", *proposal.Values.Now)
	}
	if proposal.Values.Eternal != nil {
		t.Errorf("eternal = %v, vill ha oförändrad", *proposal.Values.Eternal)
	}
	if proposal.Values.Agent != nil {
		t.Errorf("agent = %v, vill ha oförändrad", *proposal.Values.Agent)
	}
}

// TestBuildHopparÖverOgiltigaVärden kontrollerar att värden som inte
// går att tolka lämnas borta - och att det står i noteringarna.
func TestBuildHopparÖverOgiltigaVärden(t *testing.T) {
	proposal := Build(Context{
		Goal: planclient.Goal{ID: 42, Title: "Ett mål"},
		Definition: definer.Definition{
			Title:            "Ett mål",
			Description:      "Beskrivning",
			Due:              "2026-12-32",
			Icon:             "inte en ikon",
			RecurrenceRule:   "INTERVAL=0;FREQ=DAILY",
			EstimatedMinutes: -5,
		},
		NextPhase: "plan",
	})

	if proposal.Values.Due != nil {
		t.Errorf("due = %v, vill ha hoppat över", *proposal.Values.Due)
	}
	if proposal.Values.Icon != nil {
		t.Errorf("ikon = %v, vill ha hoppat över", *proposal.Values.Icon)
	}
	if proposal.Values.RecurrenceRule != nil {
		t.Errorf("upprepningsregel = %v, vill ha hoppat över", *proposal.Values.RecurrenceRule)
	}
	if proposal.Values.EstimatedMinutes != nil {
		t.Errorf("uppskattning = %v, vill ha hoppat över", *proposal.Values.EstimatedMinutes)
	}
	if len(proposal.Notes) < 3 {
		t.Errorf("noteringar = %v, vill ha en för varje hoppat över värde", proposal.Notes)
	}
}

// TestBuildBekräftarPlatsNamn kontrollerar att bara platser som finns
// föreslås - och att målets egna platser finns kvar.
func TestBuildBekräftarPlatsNamn(t *testing.T) {
	proposal := Build(Context{
		Goal: planclient.Goal{
			ID:     42,
			Title:  "Ett mål",
			Places: []planclient.Place{{ID: 1, Name: "Hem"}},
		},
		Definition: definer.Definition{
			Title:  "Ett mål",
			Places: []string{"Gymmet", "Kontoret"},
		},
		Places: []planclient.Place{
			{ID: 1, Name: "Hem"},
			{ID: 2, Name: "Gymmet"},
		},
		NextPhase: "plan",
	})

	if len(proposal.Values.PlaceIDs) != 2 {
		t.Fatalf("platser = %v, vill ha Hem och Gymmet", proposal.Values.PlaceIDs)
	}
	if proposal.Values.PlaceIDs[0] != 1 || proposal.Values.PlaceIDs[1] != 2 {
		t.Errorf("platser = %v", proposal.Values.PlaceIDs)
	}
}

// TestBuildFöreslårUpprepning kontrollerar att en giltig regel ger
// regel, läge och startdatum.
func TestBuildFöreslårUpprepning(t *testing.T) {
	proposal := Build(Context{
		Goal: planclient.Goal{ID: 42, Title: "Ett mål"},
		Definition: definer.Definition{
			Title:          "Ett mål",
			Due:            "2026-12-01",
			RecurrenceRule: "FREQ=WEEKLY;BYDAY=MO",
		},
		NextPhase: "plan",
	})

	if proposal.Values.RecurrenceRule == nil || *proposal.Values.RecurrenceRule != "FREQ=WEEKLY;BYDAY=MO" {
		t.Errorf("upprepningsregel = %v", proposal.Values.RecurrenceRule)
	}
	if proposal.Values.RecurrenceMode == nil || *proposal.Values.RecurrenceMode != "schedule" {
		t.Errorf("upprepningsstrategi = %v", proposal.Values.RecurrenceMode)
	}
	if proposal.Values.RecurrenceStartDate == nil || *proposal.Values.RecurrenceStartDate != "2026-12-01" {
		t.Errorf("startdatum = %v", proposal.Values.RecurrenceStartDate)
	}
}

// TestPayloadRoundTrip kontrollerar att ett förslag överlever en tur
// genom Plan - det är så exakt det godkända förslaget läses tillbaka.
func TestPayloadRoundTrip(t *testing.T) {
	proposal := Build(Context{
		Goal: planclient.Goal{ID: 42, Title: "Ett mål"},
		Definition: definer.Definition{
			Title:            "Ett nytt namn",
			Description:      "Beskrivning",
			EstimatedMinutes: 45,
		},
		NextPhase: "plan",
	})

	payload := proposal.Payload()
	if payload["kind"] != Kind {
		t.Errorf("typ = %v, vill ha %q", payload["kind"], Kind)
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var läs map[string]any
	if err := json.Unmarshal(raw, &läs); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	tillbaka, err := FromPayload(läs)
	if err != nil {
		t.Fatalf("FromPayload: %v", err)
	}

	if tillbaka.Values.Title == nil || *tillbaka.Values.Title != "Ett nytt namn" {
		t.Errorf("titel = %v", tillbaka.Values.Title)
	}
	if tillbaka.Values.EstimatedMinutes == nil || *tillbaka.Values.EstimatedMinutes != 45 {
		t.Errorf("uppskattning = %v", tillbaka.Values.EstimatedMinutes)
	}
}

// TestApplySkriverExaktGodkändaVärden kontrollerar att appliceringen
// skriver de godkända värdena - och inget annat.
func TestApplySkriverExaktGodkändaVärden(t *testing.T) {
	var goalRequest planclient.UpdateGoalRequest
	var goalTags []string
	var goalPlaces []int64
	var goalStatus int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "PATCH /api/v1/goals/42":
			_ = json.NewDecoder(r.Body).Decode(&goalRequest)
			_ = json.NewEncoder(w).Encode(map[string]planclient.Goal{"goal": {ID: 42}})
		case "PUT /api/v1/goals/42/tags":
			var request struct {
				Tags []string `json:"tags"`
			}
			_ = json.NewDecoder(r.Body).Decode(&request)
			goalTags = request.Tags
			_ = json.NewEncoder(w).Encode(request.Tags)
		case "PUT /api/v1/goals/42/places":
			var request struct {
				PlaceIDs []int64 `json:"place_ids"`
			}
			_ = json.NewDecoder(r.Body).Decode(&request)
			goalPlaces = request.PlaceIDs
			_ = json.NewEncoder(w).Encode([]planclient.Place{})
		case "GET /api/v1/statuses":
			_ = json.NewEncoder(w).Encode([]planclient.Status{
				{ID: 1, Name: "define"},
				{ID: 2, Name: "plan"},
			})
		case "PATCH /api/v1/goals/42/status":
			var request struct {
				StatusID int `json:"status_id"`
			}
			_ = json.NewDecoder(r.Body).Decode(&request)
			goalStatus = request.StatusID
			_ = json.NewEncoder(w).Encode(map[string]any{"status_id": request.StatusID})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("PLAN_API_URL", server.URL)
	t.Setenv("PLAN_API_KEY", "test-key")
	client, err := planclient.New()
	if err != nil {
		t.Fatalf("create Plan API client: %v", err)
	}

	title := "Ett nytt namn"
	tags := []string{"ute", "helg"}
	places := []int64{1, 2}
	status := "plan"

	proposal := Proposal{
		Kind:   Kind,
		GoalID: 42,
		Values: Values{
			Title:    &title,
			Tags:     tags,
			PlaceIDs: places,
			Status:   &status,
		},
	}

	if err := proposal.Apply(context.Background(), client, 42); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if goalRequest.Title == nil || *goalRequest.Title != "Ett nytt namn" {
		t.Errorf("titel = %v", goalRequest.Title)
	}
	if goalRequest.Description != nil {
		t.Errorf("beskrivning = %v, vill ha oförändrad", *goalRequest.Description)
	}
	if len(goalTags) != 2 {
		t.Errorf("taggar = %v", goalTags)
	}
	if len(goalPlaces) != 2 {
		t.Errorf("platser = %v", goalPlaces)
	}
	if goalStatus != 2 {
		t.Errorf("status = %d, vill ha 2 (plan)", goalStatus)
	}
}
