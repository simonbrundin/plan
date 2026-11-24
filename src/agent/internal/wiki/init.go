package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Markörer runt det block som wikictl init skriver in i AGENTS.md.
const (
	AgentsBlockStart = "<!-- wiki:start -->"
	AgentsBlockEnd   = "<!-- wiki:end -->"
)

// AgentsFileName är filen där wikinreglerna läggs in.
const AgentsFileName = "AGENTS.md"

// AgentsBlock är texten som läggs in i ett repos AGENTS.md.
// Den är kort - reglerna i sin helhet finns i den globala AGENTS.md.
const AgentsBlock = AgentsBlockStart + `
## Kunskaps-wiki

Det här repot delar en gemensam kunskaps-wiki med alla andra repor.
Wikin ligger i ` + "`~/wiki`" + ` och är en persistent kunskapsgraf som
agenterna underhåller - kunskapen härleds inte om vid varje fråga.

**Fem regler:**

1. **Läs wikin innan du frågar användaren.** Svaret finns ofta redan
   dokumenterat - leta först.
2. **Skriv en not i wikin inbox när något stort händer.** Incident, beslut,
   ny feature eller en icke-obvious bugg.
3. **Lägg in råkällor** du lär dig av med ` + "`wikictl source add`" + `.
   De är immutabla och ger spårbarhet.
4. **Håll wikin aktuell.** Ett påstående som en ny källa motsäger ska
   ändras, inte ligga kvar. ` + "`wikictl lint`" + ` hittar trasiga länkar,
   ödelappar och föråldrad info.
5. **Fila tillbaka värdefulla svar** som sidor, så att de inte försvinner
   i chatthistoriken.

` + "```bash" + `
wikictl search "bigram"                  # sök i hela wikin
wikictl search "fel" --project <repo>    # sök inom ett projekt
wikictl note --title "Rubrik" --tags "bug" --body "Vad som hände och varför."
wikictl page --type concept --title "Rubrik" --body "..."   # skriv en sida
wikictl link --from concepts/a --to concepts/b               # korslänka
wikictl source add --file kalla.md --title "Rubrik"         # råkälla
wikictl ingest                          # flytta inboxen in i grafen
wikictl lint                            # hälsa
` + "```" + `

Anteckningar taggas automatiskt med ` + "`projekt:<repo>`" + `.
Schemat för wikin står i ` + "`~/wiki/AGENTS.md`" + `.
` + AgentsBlockEnd

// EnsureAgentsBlock skriver wikinblocket in i ett repos AGENTS.md.
// Om filen saknas skapas den, om blocket redan finns ersätts det.
// Funktionen är idempotent: att köra den två ganger ger samma resultat.
func EnsureAgentsBlock(repoDir string) (bool, error) {
	path := filepath.Join(repoDir, AgentsFileName)

	existing, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("kunde inte läsa %s: %w", path, err)
	}

	content := string(existing)
	updated := insertBlock(content)

	if updated == content {
		return false, nil
	}
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		return false, fmt.Errorf("kunde inte skriva %s: %w", path, err)
	}
	return true, nil
}

// insertBlock sätter in wikinblocket i befintlig text.
func insertBlock(content string) string {
	start := strings.Index(content, AgentsBlockStart)
	end := strings.Index(content, AgentsBlockEnd)

	if start >= 0 && end > start {
		return content[:start] + AgentsBlock + content[end+len(AgentsBlockEnd):]
	}

	if strings.TrimSpace(content) == "" {
		return "# Agentinstruktioner\n\n" + AgentsBlock + "\n"
	}

	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return content + "\n" + AgentsBlock + "\n"
}
