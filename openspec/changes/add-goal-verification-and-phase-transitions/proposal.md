## Why

Plan har redan mål, de fem faserna (`define`, `plan`, `execute`, `verify`,
`completed`), agentkörningar, godkännanden, human gate och kostnadskontroll.
Men agenten kan idag bara ta ett mål från `define` till `plan` - kodprofilens
`coder` gör sedan kodändringar. `profile.Life` finns som teknisk policy men har
ingen exekveringsmotor bakom sig.

Det som saknas är just det användaren efterfrågar: att en agent ska kunna ta
ett mål från definition till **slutfört och verifierat**, oavsett om målet
gäller kod eller något annat i livet.

Två saker saknas för att det ska gå:

1. **Fasövergångarna är inte styrda.** Ett måls fas ändras via en generell
   statusendpoint. Det är modellen som bestämmer vad som händer, inte
   koden. Därför är faser i praktiken tafelexempel.
2. **"Verifierat" finns inte som data.** `agent_approvals` har ett fält
   `verification_plan`, men inget lagrar *vad som faktiskt styrker* att ett mål
   är klart. Utan det kan agenten föreslå `completed` utan bevis, och då är
   "verifierad" bara en sträng i ett sammandrag.

## What Changes

- **Fasmaskin i Plan API.** Ett nytt anrop `POST /api/v1/goals/:id/phase`
  flyttar ett mål mellan faser. Varje övergång har villkor som API:t
  kontrollerar, och övergången loggas i `status_updates`. Agenten kan inte
  skriva målets status fritt - villkoren gäller likadant för agent och
  människa.
- **"Jag är klar" är en förstaklassig väg.** Ett mål ska kunna avslutas av
  människan med en enda handling, utan verifieringsplan och utan att agenten
  varit inblandad. Exempel: målet "Byt glödlampa" avslutas när användaren
  säger att den är bytt. En verifieringsplan är ett *extra*, inte ett krav.
- **Verifieringsplan är valfri.** Ett mål kan ha en verifieringsplan: typ
  (`checklista`, `delmal`, `kalender`, `bevis`, `matvarde`, `kod`), ett villkor
  och en styrka (`stark` / `svag`). När en plan finns blir den villkoret för
  att agenten får föreslå att målet är klart - och gränssnittet visar
  beviset innan du bestämmer.
- **Bevislager på målet.** Bevis (text, värde, tidpunkt, filreferens, länk
  till kalenderhändelse eller kodstatus) sparas på målet och hör till ett
  verifieringsomdöme. Beviset är kvar i Plan, så det överlever att
  agenttjänsten är nere.
- **Agentens åtgärdsprofil bestäms av uppgiften, inte av målets ämne.**
  Målet har ingen "typ". En körning har istället en profil: `planering`
  (standard - agenten pratar, bryter ned och föreslår, men skriver ingenting),
  `research` (läser och samlar underlag) eller `arbete` (arbetar i en isolerad
  kopia och föreslår ändringar). Agenten föreslår profil i definitionen och
  användaren godkänner den. Målets *ämne* - kod, hälsa, ekonomi - är i stället
  en tagg, vilket Plan redan har.
- **Godkännande: en enda regel.** Agentens förslag för ett mål godkänns av
  dig innan agenten får agera på målet. Allt som lämnar datorn - köpa, boka,
  skicka, ringa - kräver alltid ett eget godkännande, oavsett inställning.
  En per-mål-inställning kan dessutom kräva godkännande för varje åtgärd.
- **Enhetlig målterminologi.** Allt agenten skapar är ett *mål*. Det finns
  ingen separat "checkpoint"-modell i målmodellen - agentens plan är ett träd
  av mål, där undermål är arbetsenheten. Det är samma begrepp användaren
  redan använder.

## Impact

- Affected specs:
  - `goal-phase-transitions` (ny)
  - `goal-verification` (ny)
  - `agent-assisted-goals` (utökas)
- Affected code:
  - `environments/base/db/schema.yaml` - `goal_verifications`,
    `goal_evidence`, `goal_verification_results`, `goals.approval_mode`
  - `src/api/internal/models/goal.go`, `src/api/internal/models/agent_run.go`
  - `src/api/internal/handlers/goals.go` - ny fasendpoint
  - `src/api/internal/handlers/` - verifierings- och bevisendpoints
  - `src/api/internal/middleware/agent_scope.go` - agentens scope måste
    kunna läsa och skriva verifieringar men inte besluta om dem
  - `src/contracts/plan-agent.yaml` - nya anrop och typer
  - `src/agent/internal/goaldefinition` - föreslår verifieringsplan och
    åtgärdsprofil
  - `src/agent/internal/agentworker` - driver faser mot de nya villkoren
  - `src/nuxt` - visa verifieringsplan, bevis och godkännande i agent-docken
- Inte i den här ändringen: själva agent-rollerna (definera/planera/utföra/
  verifiera), morgonpuls, kvällscheck-in och notiser. De bygger på den här
  ändringen och kommer i egna changes.
