## 1. Datamodell

- [ ] 1.1 `goals.approval_mode` (`plan` | `varje_atgard`, default `plan`) i
      `environments/base/db/schema.yaml`
- [ ] 1.2 Tabell `goal_verifications` (mål, typ, villkor, styrka, skapad)
- [ ] 1.3 Tabell `goal_evidence` (mål, verifiering, typ, värde, tidsstämpel,
      källa)
- [ ] 1.4 Tabell `goal_verification_results` (mål, verifiering, utfall,
      godkänd-av, godkänd-am, orsak)
- [ ] 1.5 Kolumn för hur ett mål avslutades (`deklarerat` | `verifierat`)
- [ ] 1.6 Idempotent migrering i `src/api/internal/database/db.go` - samma
      mönster som tidigare statusmigreringar

## 2. Fasmaskin i Plan API

- [ ] 2.1 `POST /api/v1/goals/:id/phase` med body `{to, override_reason?}`
- [ ] 2.2 Villkor per övergång i `src/api/internal/handlers/goals.go` (eller
      ny `phase.go`) - `define`→`plan`, `plan`→`execute`,
      `execute`→`verify`
- [ ] 2.3 `POST /api/v1/goals/:id/complete` - genväg för människan: avslutar
      målet oavsett fas, registrerar `deklarerat`/`verifierat`
- [ ] 2.4 409 med `reason` och saknat villkor i svaret
- [ ] 2.5 Loggning i `status_updates` (från, till, vem) + avvikelseevent vid
      override och bakåtgång
- [ ] 2.6 Befintliga `PATCH /api/v1/goals/:id/status` går via samma maskin
- [ ] 2.7 Enhetstester för varje villkor, inklusive agent-scope

## 3. Verifiering och bevis

- [ ] 3.1 `GET/PUT /api/v1/goals/:id/verification` - valfri plan
- [ ] 3.2 `POST /api/v1/goals/:id/evidence`
- [ ] 3.3 `POST /api/v1/goals/:id/verification/run` - kör verifieringen och
      ger omdöme
- [ ] 3.4 `POST /api/v1/verification-results/:id/approve` - kräver
      människasession, agent-scope får 403
- [ ] 3.5 Ogiltigförklarande av omdöme när verifieringsplanen ändras
- [ ] 3.6 Agentens förslag att avsluta kräver godkänd verifiering NÄR målet
      har en plan; annars går förslaget fram men markeras "utan bevis"

## 4. Agenten

- [ ] 4.1 `goaldefinition` föreslår verifieringsplan (typ, villkor, styrka)
      och åtgärdsprofil i definitionsförslaget
- [ ] 4.2 `agentworker` härleder åtgärdsprofilen ur det godkända förslaget
- [ ] 4.3 Ny profil-policy: `planering` och `research` mappar mot livsprofilens
      rättigheter, `arbete` mot kodprofilens
- [ ] 4.4 `agentworker` går via fasendpointen och respekterar 409
- [ ] 4.5 `planclient` får anrop för verifiering, bevis och omdömen

## 5. Kontrakt

- [ ] 5.1 `src/contracts/plan-agent.yaml` - nya typer, anrop och
      fas-villkor; bump version
- [ ] 5.2 Plan testar att varje anrop i kontraktet finns i OpenAPI-specen och
      är tillåtet för agent-scope
- [ ] 5.3 Agenten testar att dess klientmatchar kontraktet

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
