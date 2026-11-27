## 1. Databas

- [ ] 1.1 `environments/base/db/schema.yaml`: `goals.criteria` (TEXT, nullable)
      och tabellen `goal_definition_applies`
      (id, goal_id FK CASCADE, field, from_value, to_value, decided_by,
      decided_at)
- [ ] 1.2 `environments/pr/template/db/schema.sql`: samma ändring
- [ ] 1.3 `src/api/internal/database/db.go`: idempotenta migreringar
      (`ALTER TABLE goals ADD COLUMN IF NOT EXISTS criteria TEXT` och
      `CREATE TABLE IF NOT EXISTS goal_definition_applies` + index på goal_id)

## 2. Gemensam definitionspolicy

- [ ] 2.1 Flytta `src/agent/internal/goaldefinition` till `src/internal/goaldefinition`
      (Go-paketet `plan/internal/goaldefinition`) - flytten ska inte ändra
      beteendet
- [ ] 2.2 `src/agent/internal/goaldefinition`: tunn wrapper eller borttaget -
      agenten importerar det gemensamma paketet
- [ ] 2.3 Flytta `src/agent/internal/definer` på samma sätt, så att prompten
      och JSON-formatet finns på ett ställe
- [ ] 2.4 Enhetstester flyttas med och fortsätter gå gråna

## 3. Go: modeller och LLM-klient i Plan API:t

- [ ] 3.1 `src/api/internal/llm/`: liten Anthropic-kompatibel klient
      (modell, env-nyckel `ANTHROPIC_API_KEY` eller `MINIMAX_API_KEY` +
      `MINIMAX_BASE_URL`), `Chat(ctx, prompt) (string, error)`
- [ ] 3.2 `src/api/internal/models/goal_definition.go`: `DefinitionProposal`
      (kind, goal_id, fields[], values, notes[]), `DefinitionField`
      (key, label, from, to) och `DefinitionApplyRequest`
      (fields[] = exakt de godkända fälten)
- [ ] 3.3 `src/api/internal/models/goal.go`: `Criteria` på `Goal`,
      `CreateGoalRequest`/`UpdateGoalRequest` med `criteria`
- [ ] 3.4 `src/api/internal/repository/goal_repository.go`: `criteria` i svar,
      uppdatering av `criteria`
- [ ] 3.5 `src/api/internal/repository/goal_definition_repository.go`:
      logga appliceringar, lista loggen per mål

## 4. Go: endpoints

- [ ] 4.1 `handlers/goal_definitions.go`:
      `POST /api/v1/goals/:id/definition/propose`
      - byggerluckor ur målets nuvarande värden
      - anropar LLM:en via det gemensamma paketet
      - returnerar `DefinitionProposal` med fält (från -> till) och anteckningar
      - skriver INGET till målet
      - svarar tydligt med 503 om ingen API-nyckel finns
- [ ] 4.2 `handlers/goal_definitions.go`:
      `POST /api/v1/goals/:id/definition/apply`
      - tar emot exakt de godkända fälten
      - skriver bara dessa värden
      - loggar i `goal_definition_applies`
      - kräver användarsession
- [ ] 4.3 `GET /api/v1/goals/:id/definition/applies` - loggen för målet
- [ ] 4.4 `PUT /api/v1/goals/:id/criteria` - sätt kriteriet (eller via
      befintliga måluppdateringen)
- [ ] 4.5 Routes i `src/api/cmd/api/main.go` + Swagger-kommentarer
- [ ] 4.6 Enhetstester: luckinventering (inga befintliga värden föreslås),
      validering av apply-payload (okända fält avvisas), att felaktiga värden
      avvisas

## 5. Nuxt: panelen

- [ ] 5.1 `app/types/goalDefinition.ts`: typer för förslag och fält
- [ ] 5.2 `app/composables/useGoalDefinition.ts`: `propose`, `apply`, `criteria`,
      tillstånd och fel
- [ ] 5.3 `app/components/DefinitionPanel.vue`: visar målet, markerar saknade
      fält, visar `från -> till` per fält, knappar för acceptera/ändra/avvisa
      per fält, samt kriterie-fältet
- [ ] 5.4 Panelen öppnas från målsidan med `Space+f` (leader-kommando "f")
- [ ] 5.5 Panelen öppnas från inboxen ("Bearbeta" - knappen är död kod idag)
- [ ] 5.6 Varning på målsidan: "N fält saknas" och "inget verifieringskriterium"
      - varningen försvinner när luckorna är fyllda
- [ ] 5.7 Kriteriet redigeras på målsidan (tangent, t.ex. `Space+c`)
- [ ] 5.8 `HelpModal.vue`: dokumentera `Space+f` och kriterietangenten

## 6. Verifiering och dokumentation

- [ ] 6.1 Go: `go build ./...`, `go test ./...`, `go vet ./...`
- [ ] 6.2 Nuxt: `bun run build`, `bun test`, `bun run test`
- [ ] 6.3 Manuell verifiering i dev: panel från inbox och målsida, förslag med
      och utan API-nyckel, fältvis godkännande (bara ett fält appliceras),
      avvisat fält ändras inte, kriterium sparas, fasbytet går alltid
- [ ] 6.4 Wiki-not om definitionsassistenten
