package planclient

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// anrop är ett API-anrop agenten gör mot Plan.
type anrop struct {
	Method string
	Path   string
}

// uttryck för hur klienten sätter ihop sökvägar.
var (
	// anropSökning hittar anropen i klientens källkod.
	anropSökning = regexp.MustCompile(`do\(ctx,\s*(http\.Method[A-Za-z]+),\s*([^,)]+(?:\([^()]*\))?)`)

	// sökvägKonstant hittar sökvägskonstanterna i paths.go.
	sökvägKonstant = regexp.MustCompile(`(path[A-Za-z]+)\s*=\s*"([^"]+)"`)

	// variabelSökning hittar lokala variabler som sätts till en sökväg.
	variabelSökning = regexp.MustCompile(`(\w+)\s*:?=\s*(path[A-Za-z]+)`)
)

// klientanrop går igenom klientens källkod och listar varje anrop den gör.
//
// Klienten går genom do() för allt som når Plan, så det räcker att läsa
// de anropen. En ny sökväg som inte gissas av skannern gör att testet
// misslyckas - hellre det än att en ny sökväg smyger sig förbi kontraktet.
func klientanrop(t *testing.T) []anrop {
	t.Helper()

	filer, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("kan inte lista klientens filer: %v", err)
	}

	konstanter := map[string]string{}
	for _, fil := range filer {
		if strings.HasSuffix(fil, "_test.go") {
			continue
		}

		källa, err := os.ReadFile(fil)
		if err != nil {
			t.Fatalf("kan inte läsa %s: %v", fil, err)
		}

		for _, match := range sökvägKonstant.FindAllStringSubmatch(string(källa), -1) {
			konstanter[match[1]] = match[2]
		}
	}

	var samlade []anrop
	for _, fil := range filer {
		if strings.HasSuffix(fil, "_test.go") {
			continue
		}

		källa, err := os.ReadFile(fil)
		if err != nil {
			t.Fatalf("kan inte läsa %s: %v", fil, err)
		}

		text := string(källa)

		// Lokala variabler som håller en sökväg ska kunna lösas upp,
		// annars vet inte skannern vilken väg anropet går till.
		variabler := map[string]string{}
		for _, match := range variabelSökning.FindAllStringSubmatch(text, -1) {
			variabler[match[1]] = match[2]
		}

		for _, match := range anropSökning.FindAllStringSubmatch(text, -1) {
			metod := strings.ToUpper(strings.TrimPrefix(match[1], "http.Method"))
			sökväg, fel := lösSökväg(match[2], konstanter, variabler)

			if fel != "" {
				t.Errorf("kan inte tolka sökvägen %q i %s: %s", match[2], fil, fel)
				continue
			}

			samlade = append(samlade, anrop{Method: metod, Path: sökväg})
		}
	}

	return samlade
}

// lösSökväg översätter ett uttryck i klientens källkod till kontraktets
// form - "/api/v1/goals/{id}".
func lösSökväg(uttryck string, konstanter, variabler map[string]string) (string, string) {
	uttryck = strings.TrimSpace(uttryck)

	// fmt.Sprintf(pathX, id) - parametern skrivs {id} i kontraktet.
	if match := regexp.MustCompile(`^fmt\.Sprintf\((path[A-Za-z]+)`).FindStringSubmatch(uttryck); match != nil {
		return normalisera(konstanter[match[1]]), ""
	}

	// pathX - en konstant, eventuellt med en fråga på slutet.
	if match := regexp.MustCompile(`^(path[A-Za-z]+)`).FindStringSubmatch(uttryck); match != nil {
		return normalisera(konstanter[match[1]]), ""
	}

	// En lokal variabel som pekar på en sökväg.
	if match := regexp.MustCompile(`^(\w+)$`).FindStringSubmatch(uttryck); match != nil {
		if sökväg, ok := variabler[match[1]]; ok {
			return normalisera(konstanter[sökväg]), ""
		}
		return "", "variabeln pekar inte på någon känd sökväg"
	}

	return "", "okänt uttryck"
}

// normalisera gör om en sökväg till kontraktets form: %d blir {id} och
// frågan efter ? tas bort.
func normalisera(sökväg string) string {
	sökväg = strings.SplitN(sökväg, "?", 2)[0]
	return strings.ReplaceAll(sökväg, "%d", "{id}")
}

// TestKlientensAnropFinnsIKontraktet kontrollerar att varje anrop
// agentens klient gör är markerat som agentens eget anrop i kontraktet.
//
// Plan tillåter bara de anropen för agent-scope. Ett anrop som inte är
// markerat ger 403 i prod, där agentens nyckel har scope 'agent'. I dev
// har nyckeln scope 'full', så felet syns inte där - testet är därför
// enda vägen att upptäcka det innan prod.
func TestKlientensAnropFinnsIKontraktet(t *testing.T) {
	k := läsKontrakt(t)

	markerade := map[string]bool{}
	for _, e := range k.Endpoints {
		if e.Agent {
			markerade[e.Method+" "+e.Path] = true
		}
	}

	anropen := klientanrop(t)
	if len(anropen) == 0 {
		t.Fatal("skannern hittade inga anrop - kontrollera att den följer med klientens ändringar")
	}

	for _, a := range anropen {
		if !markerade[a.Method+" "+a.Path] {
			t.Errorf("klienten gör %s %s, men anropet är inte markerat som agentens eget i kontraktet", a.Method, a.Path)
		}
	}
}

// TestKlientensAnropFinnsIBeskrivningen kontrollerar att varje anrop
// agenten gör finns bland kontraktets endpoints - inte bara bland de
// markerade.
func TestKlientensAnropFinnsIBeskrivningen(t *testing.T) {
	k := läsKontrakt(t)

	kända := map[string]bool{}
	for _, e := range k.Endpoints {
		kända[e.Method+" "+e.Path] = true
	}

	for _, a := range klientanrop(t) {
		if !kända[a.Method+" "+a.Path] {
			t.Errorf("klienten gör %s %s, men anropet finns inte i kontraktet", a.Method, a.Path)
		}
	}
}

var _ = fmt.Sprintf
