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

// TestBuildFöreslårVerifieringsplan kontrollerar att agentens förslag
// innehåller verifieringsplanen - den är det som gör "verifierat" till
// bevis i stället för förklaringsmodeller.
func TestBuildFöreslårVerifieringsplan(t *testing.T) {
	proposal := Build(Context{
		Goal: planclient.Goal{ID: 42, Title: "Byt glödlampa"},
		Definition: definer.Definition{
			Title:                 "Byt glödlampa",
			VerificationType:      "checklista",
			VerificationCriterion: "Tre lampor bytta",
			VerificationStrength:  "svag",
		},
	})

	if proposal.Values.VerificationType == nil || *proposal.Values.VerificationType != "checklista" {
		t.Fatalf("verifieringstyp = %v", proposal.Values.VerificationType)
	}
	if proposal.Values.VerificationCriterion == nil || *proposal.Values.VerificationCriterion != "Tre lampor bytta" {
		t.Errorf("villkor = %v", proposal.Values.VerificationCriterion)
	}
	if proposal.Values.VerificationStrength == nil || *proposal.Values.VerificationStrength != "svag" {
		t.Errorf("styrka = %v", proposal.Values.VerificationStrength)
	}
}

// TestBuildHopparÖverTrasigVerifieringsplan kontrollerar att en plan
// agenten inte kan stå för sig inte föreslås.
//
// Ett förslag med en trasig plan är värre än inget förslag - människan
// skulle godkänna något som inte fungerar.
func TestBuildHopparÖverTrasigVerifieringsplan(t *testing.T) {
	proposal := Build(Context{
		Goal: planclient.Goal{ID: 42, Title: "Byt glödlampa"},
		Definition: definer.Definition{
			VerificationType:      "troligen",
			VerificationCriterion: "Tre lampor bytta",
		},
	})

	if proposal.Values.VerificationType != nil {
		t.Errorf("en okänd typ ska inte föreslås, fick %v", *proposal.Values.VerificationType)
	}
	if len(proposal.Notes) == 0 {
		t.Error("människan ska få en notering om att planen hoppades över")
	}
}

// TestBuildHopparÖverPlanUtanVillkor kontrollerar att en verifiering utan
// villkor inte föreslås.
func TestBuildHopparÖverPlanUtanVillkor(t *testing.T) {
	proposal := Build(Context{
		Goal: planclient.Goal{ID: 42, Title: "Byt glödlampa"},
		Definition: definer.Definition{
			VerificationType: "bevis",
		},
	})

	if proposal.Values.VerificationType != nil {
		t.Error("en plan utan villkor ska inte föreslås")
	}
}

// TestBuildFöreslårÅtgärdsprofil kontrollerar att profilen kommer med i
// förslaget - den bestämmer vad agenten får göra i körningen.
func TestBuildFöreslårÅtgärdsprofil(t *testing.T) {
	proposal := Build(Context{
		Goal: planclient.Goal{ID: 42, Title: "Skriv en rapport"},
		Definition: definer.Definition{
			ActionProfile: "research",
		},
	})

	if proposal.Values.ActionProfile == nil || *proposal.Values.ActionProfile != "research" {
		t.Fatalf("åtgärdsprofil = %v", proposal.Values.ActionProfile)
	}
}

// TestBuildHopparÖverOkändÅtgärdsprofil kontrollerar att en okänd profil
// inte går med i förslaget.
func TestBuildHopparÖverOkändÅtgärdsprofil(t *testing.T) {
	proposal := Build(Context{
		Goal: planclient.Goal{ID: 42, Title: "Skriv en rapport"},
		Definition: definer.Definition{
			ActionProfile: "gör vad du vill",
		},
	})

	if proposal.Values.ActionProfile != nil {
		t.Errorf("en okänd profil ska inte föreslås, fick %v", *proposal.Values.ActionProfile)
	}
	if len(proposal.Notes) == 0 {
		t.Error("människan ska få en notering om den okända profilen")
	}
}

// TestApplySätterVerifieringsplanen kontrollerar att den godkända planen
// sätts på målet - och bara den.
func TestApplySätterVerifieringsplanen(t *testing.T) {
	var sattTyp, sattVillkor, sattStyrka string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && r.URL.Path == "/api/v1/goals/42/verification" {
			var request struct {
				Type      string `json:"type"`
				Criterion string `json:"criterion"`
				Strength  string `json:"strength"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Errorf("decode verification: %v", err)
			}
			sattTyp = request.Type
			sattVillkor = request.Criterion
			sattStyrka = request.Strength
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	t.Setenv("PLAN_API_URL", server.URL)
	t.Setenv("PLAN_API_KEY", "test-key")
	client, err := planclient.New()
	if err != nil {
		t.Fatalf("create Plan API client: %v", err)
	}

	typ := "checklista"
	villkor := "Tre lampor bytta"
	styrka := "svag"
	proposal := Proposal{
		Values: Values{
			VerificationType:      &typ,
			VerificationCriterion: &villkor,
			VerificationStrength:  &styrka,
		},
	}

	if err := proposal.Apply(context.Background(), client, 42); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if sattTyp != "checklista" || sattVillkor != "Tre lampor bytta" || sattStyrka != "svag" {
		t.Errorf("planen som sattes = %q / %q / %q", sattTyp, sattVillkor, sattStyrka)
	}
}

// TestApplyUtanVerifieringsplanSätterIngen kontrollerar att ett förslag
// utan plan inte får en plan i efterhand.
func TestApplyUtanVerifieringsplanSätterIngen(t *testing.T) {
	var anrop bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/goals/42/verification" {
			anrop = true
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	t.Setenv("PLAN_API_URL", server.URL)
	t.Setenv("PLAN_API_KEY", "test-key")
	client, err := planclient.New()
	if err != nil {
		t.Fatalf("create Plan API client: %v", err)
	}

	proposal := Proposal{Values: Values{}}
	if err := proposal.Apply(context.Background(), client, 42); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if anrop {
		t.Error("en verifieringsplan ska inte sättas utan att den ingick i förslaget")
	}
}
