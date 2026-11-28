package phase

import "testing"

// TestEvaluateAgentTransitions kontrollerar att agenten bara får gå vidare
// när villkoret för övergången är uppfyllt.
func TestEvaluateAgentTransitions(t *testing.T) {
	cases := []struct {
		name     string
		from     string
		to       string
		facts    Facts
		wantCode string
	}{
		{
			name:  "define till plan utan godkänd definition",
			from:  Define,
			to:    Plan,
			facts: Facts{},
			// Ett mål utan godkänd definition får inte gå vidare.
			wantCode: ReasonDefinitionNotApproved,
		},
		{
			name:     "define till plan med godkänd definition",
			from:     Define,
			to:       Plan,
			facts:    Facts{HasApprovedDefinition: true},
			wantCode: "",
		},
		{
			name:     "plan till execute utan delmål",
			from:     Plan,
			to:       Execute,
			facts:    Facts{HasApprovedDefinition: true},
			wantCode: ReasonMissingPlannedChildren,
		},
		{
			name: "plan till execute med planerade delmål",
			from: Plan,
			to:   Execute,
			facts: Facts{
				HasApprovedDefinition: true,
				HasPlannedChildren:    true,
			},
			wantCode: "",
		},
		{
			name: "execute till verify med olöst delmål",
			from: Execute,
			to:   Verify,
			facts: Facts{
				HasApprovedDefinition: true,
				HasPlannedChildren:    true,
				AllChildrenResolved:   false,
			},
			wantCode: ReasonUnresolvedChildren,
		},
		{
			name: "execute till verify med lösta delmål",
			from: Execute,
			to:   Verify,
			facts: Facts{
				HasApprovedDefinition: true,
				HasPlannedChildren:    true,
				AllChildrenResolved:   true,
			},
			wantCode: "",
		},
		{
			name:     "agenten får inte avsluta ett mål",
			from:     Verify,
			to:       Completed,
			facts:    Facts{VerificationPassed: true, HasVerificationPlan: true},
			wantCode: ReasonApprovalRequired,
		},
		{
			name: "agenten får inte hoppa över faser",
			from: Define,
			to:   Execute,
			facts: Facts{
				HasApprovedDefinition: true,
				HasPlannedChildren:    true,
			},
			wantCode: ReasonPhaseSkipped,
		},
		{
			name:     "agenten får inte gå bakåt",
			from:     Execute,
			to:       Plan,
			facts:    Facts{HasApprovedDefinition: true},
			wantCode: ReasonBackwards,
		},
		{
			name:     "okänd fas",
			from:     Define,
			to:       "klar",
			facts:    Facts{},
			wantCode: ReasonUnknownPhase,
		},
		{
			name:     "okänd frånfas",
			from:     "start",
			to:       Plan,
			facts:    Facts{},
			wantCode: ReasonUnknownPhase,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Evaluate(tc.from, tc.to, tc.facts)

			if tc.wantCode == "" {
				if !result.Allowed {
					t.Fatalf("övergången skulle tillåtas men stoppades: %s (%s)", result.Code, result.Message)
				}
				return
			}

			if result.Allowed {
				t.Fatalf("övergången skulle stoppas med %q men tilläts", tc.wantCode)
			}
			if result.Code != tc.wantCode {
				t.Errorf("skäl = %q, vill ha %q", result.Code, tc.wantCode)
			}
		})
	}
}

// TestEvaluateHumanCompletes kontrollerar genvägen "jag är klar".
func TestEvaluateHumanCompletes(t *testing.T) {
	cases := []struct {
		name     string
		from     string
		facts    Facts
		wantKind string
	}{
		{
			name:     "enkelt mål utan verifiering",
			from:     Define,
			facts:    Facts{IsHuman: true},
			wantKind: CompletionDeclared,
		},
		{
			name:     "mål med godkänd verifiering",
			from:     Verify,
			facts:    Facts{IsHuman: true, HasVerificationPlan: true, VerificationPassed: true},
			wantKind: CompletionVerified,
		},
		{
			name:     "mål med verifiering som inte gått igenom",
			from:     Verify,
			facts:    Facts{IsHuman: true, HasVerificationPlan: true, VerificationPassed: false},
			wantKind: CompletionDeclared,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Evaluate(tc.from, Completed, tc.facts)

			if !result.Allowed {
				t.Fatalf("människan ska kunna avsluta målet: %s (%s)", result.Code, result.Message)
			}
			if result.CompletionKind != tc.wantKind {
				t.Errorf("avslutandet = %q, vill ha %q", result.CompletionKind, tc.wantKind)
			}
		})
	}
}

// TestEvaluateHumanMayGoBackwards kontrollerar att en människa får öppna
// upp ett mål igen - men att det räknas som en tvingad övergång.
func TestEvaluateHumanMayGoBackwards(t *testing.T) {
	result := Evaluate(Execute, Plan, Facts{IsHuman: true})

	if !result.Allowed {
		t.Fatalf("människan ska kunna gå bakåt: %s", result.Message)
	}
	if !result.Forced {
		t.Error("en bakåtgång ska markeras som tvingad")
	}
}

// TestEvaluateHumanMaySkipPhases kontrollerar att en människa får hoppa
// över faser - och att hoppet syns i svaret.
func TestEvaluateHumanMaySkipPhases(t *testing.T) {
	result := Evaluate(Define, Execute, Facts{IsHuman: true})

	if !result.Allowed {
		t.Fatalf("människan ska kunna hoppa över faser: %s", result.Message)
	}
	if !result.Forced {
		t.Error("ett fashopp ska markeras som tvingat")
	}
	if result.SkippedPhases != 1 {
		t.Errorf("hoppade faser = %d, vill ha 1", result.SkippedPhases)
	}
}

// TestEvaluateSamePhase kontrollerar att en övergång till samma fas är
// tillåten men inte gör något.
func TestEvaluateSamePhase(t *testing.T) {
	result := Evaluate(Plan, Plan, Facts{})

	if !result.Allowed {
		t.Fatal("övergång till samma fas ska tillåtas")
	}
	if result.Forced {
		t.Error("övergång till samma fas ska inte räknas som tvingad")
	}
}

// TestIndex kontrollerar fasordningen.
func TestIndex(t *testing.T) {
	if Index(Define) != 0 || Index(Completed) != len(Order)-1 {
		t.Error("fasordningen är fel")
	}
	if Index("klar") != -1 {
		t.Error("en okänd fas ska ge -1")
	}
}
