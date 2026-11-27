package profile

import "testing"

// TestAtgardsprofilerMappedRatt kontrollerar att åtgärdsprofilerna ger
// de rättigheter de ska.
//
// Det här är kärnan i profilpolitiken: planering och research får inte
// röra kod, och arbete får det. En bugg här ger agenten mer rättigheter
// än användaren godkänt.
func TestAtgardsprofilerMappedRatt(t *testing.T) {
	planering, err := ActionPolicy(ActionPlanering)
	if err != nil {
		t.Fatalf("planering: %v", err)
	}
	if planering.MayChangeCode {
		t.Error("planering får inte ändra kod")
	}
	if planering.NeedsRepo {
		t.Error("planering ska inte behöva repot")
	}
	if !planering.NeedsHumanGate {
		t.Error("planering ska alltid kräva ett mänskligt beslut")
	}

	research, err := ActionPolicy(ActionResearch)
	if err != nil {
		t.Fatalf("research: %v", err)
	}
	if research.MayChangeCode {
		t.Error("research får inte ändra kod")
	}
	if !research.NeedsHumanGate {
		t.Error("research ska alltid kräva ett mänskligt beslut")
	}

	arbete, err := ActionPolicy(ActionArbete)
	if err != nil {
		t.Fatalf("arbete: %v", err)
	}
	if !arbete.MayChangeCode {
		t.Error("arbete får ändra kod")
	}
	if !arbete.NeedsRepo {
		t.Error("arbete ska klona repot")
	}
}

// TestOkandAtgardsprofilGerStrangast kontrollerar att en okänd profil
// inte ger agenten mer rättigheter än avsett.
func TestOkandAtgardsprofilGerStrangast(t *testing.T) {
	policy, err := ActionPolicy("skriv kod åt mig")
	if err == nil {
		t.Fatal("en okänd profil ska ge ett fel")
	}
	if policy.MayChangeCode {
		t.Error("en okänd profil får inte ge kodrättigheter")
	}
}

// TestTomAtgardsprofilGerPlanering kontrollerar standardprofilen.
//
// Utan profil ska agenten få minst möjliga rättigheter - standarden är
// den strängaste, inte den vidare.
func TestTomAtgardsprofilGerPlanering(t *testing.T) {
	policy, err := ActionPolicy("")
	if err != nil {
		t.Fatalf("tom profil: %v", err)
	}
	if policy.Action != ActionPlanering {
		t.Errorf("tom profil ger %q, vill ha %q", policy.Action, ActionPlanering)
	}
	if policy.MayChangeCode {
		t.Error("tom profil får inte ge kodrättigheter")
	}
}

// TestAtgardsprofilernaFinnsIHemliga kontrollerar att varje profil i
// namnlistan går att slå upp.
func TestAtgardsprofilernaFinnsIHemliga(t *testing.T) {
	for _, namn := range ActionNames() {
		if !ActionKnown(namn) {
			t.Errorf("profilen %q finns i listan men går inte att slå upp", namn)
		}
	}
}
