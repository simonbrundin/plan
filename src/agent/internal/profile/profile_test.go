package profile

import "testing"

// TestPolicyForKännerProfilerna kontrollerar att profilerna går att slå upp.
func TestPolicyForKännerProfilerna(t *testing.T) {
	for _, name := range Names() {
		if _, err := PolicyFor(name); err != nil {
			t.Errorf("profilen %q kunde inte slås upp: %v", name, err)
		}
	}
}

// TestPolicyForOkändProfil ger kodprofilen kontrollerar att en okänd
// profil inte ger fler rättigheter än vad som är avsett.
func TestPolicyForOkändProfil(t *testing.T) {
	if _, err := PolicyFor("superagent"); err == nil {
		t.Error("en okänd profil ska ge ett fel")
	}
}

// TestPolicyForTomProfilGerKodprofilen kontrollerar att en saknad profil
// betyder kodprofilen.
func TestPolicyForTomProfilGerKodprofilen(t *testing.T) {
	policy, err := PolicyFor("")
	if err != nil {
		t.Fatalf("PolicyFor: %v", err)
	}
	if policy.Name != Coding {
		t.Errorf("profil = %q, vill ha %q", policy.Name, Coding)
	}
}

// TestLivsprofilenFårInteRöraKod kontrollerar att livsprofilen är
// begränsad utanför prompten.
func TestLivsprofilenFårInteRöraKod(t *testing.T) {
	policy, err := PolicyFor(Life)
	if err != nil {
		t.Fatalf("PolicyFor: %v", err)
	}

	if policy.MayChangeCode {
		t.Error("livsprofilen får inte ändra kod")
	}
	if policy.NeedsRepo {
		t.Error("livsprofilen ska inte behöva repot")
	}
	if !policy.NeedsHumanGate {
		t.Error("livsprofilen ska kräva ett mänskligt beslut")
	}
}

// TestLivsprofilenHarMindreResurser kontrollerar att livsprofilen får mindre
// resurser än kodprofilen.
func TestLivsprofilenHarMindreResurser(t *testing.T) {
	coding, _ := PolicyFor(Coding)
	life, _ := PolicyFor(Life)

	if life.CPULimit == coding.CPULimit {
		t.Error("livsprofilen ska ha en lägre CPU-gräns")
	}
	if life.MemoryLimit == coding.MemoryLimit {
		t.Error("livsprofilen ska ha en lägre minnesgräns")
	}
}
