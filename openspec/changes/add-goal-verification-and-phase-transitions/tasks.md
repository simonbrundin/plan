## 1. Datamodell

- [x] 1.1 `goals.approval_mode` (`plan` | `varje_atgard`, default `plan`) i
      `environments/base/db/schema.yaml`
- [x] 1.2 Tabell `goal_verifications` (mål, typ, villkor, styrka, skapad)
- [x] 1.3 Tabell `goal_evidence` (mål, verifiering, typ, värde, tidsstämpel,
      källa)
- [x] 1.4 Tabell `goal_verification_results` (mål, verifiering, utfall,
      godkänd-av, godkänd-am, orsak)
- [x] 1.5 Kolumn för hur ett mål avslutades (`deklarerat` | `verifierat`)
- [x] 1.6 Idempotent migrering i `src/api/internal/database/db.go` - samma
      mönster som tidigare statusmigreringar

Extra: `goal_phase_events` (mål, från, till, vem, tvingad, orsak) - loggen
som svarar på "varför gjorde agenten det här?".

## 2. Fasmaskin i Plan API

- [x] 2.1 `POST /api/v1/goals/:id/phase` med body `{to, override_reason?}`
- [x] 2.2 Villkor per övergång i `src/api/internal/phase/phase.go` (ren logik,
      utan databas) - `define`→`plan`, `plan`→`execute`,
      `execute`→`verify`
- [x] 2.3 `POST /api/v1/goals/:id/complete` - genväg för människan: avslutar
      målet oavsett fas, registrerar `deklarerat`/`verifierat`
- [x] 2.4 409 med `code` (stabil skälkod) och saknat villkor i svaret
- [x] 2.5 Loggning i `status_updates` (från, till, vem) + `goal_phase_events`
      vid override, bakåtgång och fashopp
- [x] 2.6 Befintliga `PATCH /api/v1/goals/:id/status` går via samma maskin
- [x] 2.7 Enhetstester för varje villkor, inklusive agent-scope

Extra: `src/api/internal/phaseservice` - bindningen mellan maskinen och
databasen, så att både fasendpointern och statusendpointern går samma väg.

## 3. Verifiering och bevis

- [x] 3.1 `GET/PUT /api/v1/goals/:id/verification` - valfri plan
- [x] 3.2 `POST /api/v1/goals/:id/evidence`
- [x] 3.3 `POST /api/v1/goals/:id/verification/run` - kör verifieringen och
      ger omdöme
- [x] 3.4 `POST /api/v1/verification-results/:id/approve` - kräver
      människasession, agent-scope får 403
- [x] 3.5 Ogiltigförklarande av omdöme när verifieringsplanen ändras
- [x] 3.6 Agentens förslag att avsluta kräver ett mänskligt beslut - maskinen
      svarar `approval_required` (409) när en agent försöker avsluta. Blir
      målet "verifierat" krävs ett godkänt omdöme, annars "deklarerat"

## 4. Agenten

- [x] 4.1 `goaldefinition` föreslår verifieringsplan (typ, villkor, styrka)
      och åtgärdsprofil i definitionsförslaget
- [x] 4.2 `agentworker` härleder åtgärdsprofilen ur det godkända förslaget
- [x] 4.3 Ny profil-policy: `planering` och `research` mappar mot livsprofilens
      rättigheter, `arbete` mot kodprofilens
- [x] 4.4 `agentworker` går via fasendpointen och respekterar 409
- [x] 4.5 `planclient` får anrop för verifiering, bevis och omdömen

## 5. Kontrakt

- [x] 5.1 `src/contracts/plan-agent.yaml` - nya typer, anrop och
      fas-villkor; bump version
- [x] 5.2 Plan testar att varje anrop i kontraktet finns i OpenAPI-specen och
      är tillåtet för agent-scope
- [x] 5.3 Agenten testar att dess klientmatchar kontraktet

## 6. Gränssnitt

- [ ] 6.1 "Jag är klar" - avslutar ett mål med en handling, även från `define`
- [ ] 6.2 Verifieringsplan syns på målsidan - typ, villkor, styrka - och kan
      läggas till, ändras eller tas bort
- [ ] 6.3 Bevislista på målet med möjlighet att lägga till bevis
- [ ] 6.4 Omdöme syns i agent-docken med knappar godkänn/avslå
- [ ] 6.5 Agentens förslag visar profilen och beviset innan godkännande
- [ ] 6.6 Målets "avslutades som" - deklarerat eller verifierat - syns på
      målet och i listan av avslutade mål

## 7. Dokumentation och wiki

- [ ] 7.1 `docs/agent-platform.md` - fasmaskinen, verifieringen, profilerna
      och de nya anropen
- [ ] 7.2 Wikinoter om besluten (fasmaskin i Plan, verifierare som data,
      profil per uppgift)
- [ ] 7.3 `src/nuxt` in-app docs uppdateras
