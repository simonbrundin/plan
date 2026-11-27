## Why

Plan har fem målfaser (`define`, `plan`, `execute`, `verify`, `completed`) och
en agent som kan definiera mål. Men agentens definitionsmotor är hängd på
`goals.agent = true` och körs via `agent-worker` -> Kubernetes-Job -> polling.
Det är maskineri byggd för att en *kodagent ska arbeta på målet*, inte för
att *jag ska definiera ett mål*.

Följden i praktiken:

- Ett nytt mål hamnar i `define` och står där. Det finns ingen hjälp att
  fylla fasen - bara etiketten "Definiera".
- Att få hjälp kräver agentflagga, agent-API:t igång, en worker-poll och ett
  godkännande i en agent-dock. Det är tungt för en snabb, mänsklig handling.
- Godkännandet är allt-eller-inget: ett `approve` applicerar hela
  values-paketet. Det man egentligen vill godkänna är fält för fält.

Det som saknas är en **definitionsassistent** som ligger i den vanliga appen,
svarar på sekunder, fyller luckor (aldrig över), och där allt skrivs till
målet först efter att användaren godkänt varje fält.

## What Changes

- **Ny capability `goal-definition-assistant`.**
- **Definitionspanel** i gränssnittet: visas från inboxen ("Bearbeta") och
  från målsidan (`Space+f`, leader + f). Panelen visar målet som det *kommer
  att se ut* - titel, beskrivning, deadline, plats, upprepning, tidsåtgång,
  taggar - och de fält som saknas.
- **Förslag, inte svar.** LLM:en fyller ENDAST de luckor som finns, med samma
  policy som agentens `goaldefinition` idag: ett fält som redan har ett värde
  lämnas orört. Antaganden redovisas som anteckningar.
- **Fältvis godkännande.** Varje föreslaget fält visas som `från -> till` och
  kan accepteras, ändras eller avvisas separat. Godkännande kräver
  användarsession. Server applicerar EXAKT de godkända värdena - aldrig en ny
  generering.
- **Verifieringskriterium som del av definitionen.** Ett nytt fält
  `goals.criteria` (fritextlista). För ett kodmål är det i praktiken
  acceptanskriteriet - "testet är grönt" - vilket binder ihop `define`,
  `plan`, `execute` och `verify`. Fältet är fritt: det får lämnas tomt.
- **Mjuk signal, ingen gate.** `define -> plan` blockeras aldrig. Panelen
  visar bara en varning ("3 fält saknas", "inget verifieringskriterium").
- **Synkron endpoint i Plan API:t.** `POST /api/v1/goals/:id/definition/propose`
  och `POST /api/v1/goals/:id/definition/apply`. Plan API:t får en egen liten
  LLM-klient i stället för att gå via agenttjänsten - definitionen ska inte
  köa bakom en worker och ett Kubernetes-jobb.
- **Delad policy.** Definitionens policy och prompt flyttas till ett gemensamt
  paket så att Plan API:t och agenten inte får två olika definitioner av
  "fylla luckor".

## Impact

- Affected specs:
  - `goal-definition-assistant` (ny capability)
  - `goal-phase-transitions` (berörs: villkoret "godkänd definition" - se
    kopplingen nedan)
- Affected code:
  - `environments/base/db/schema.yaml` - `goals.criteria`, `goal_definition_applies`
  - `environments/pr/template/db/schema.sql` - samma
  - `src/api/internal/database/db.go` - idempotenta migreringar
  - `src/api/internal/models/goal.go` - `Criteria`, `DefinitionApplyRequest`
  - `src/api/internal/models/goal_definition.go` (ny) - `DefinitionProposal`,
    `DefinitionField`
  - `src/api/internal/handlers/goal_definitions.go` (ny) - propose/apply
  - `src/api/internal/llm/` (ny) - liten Anthropic-kompatibel klient
  - `src/internal/goaldefinition/` (ny, flyttad från `src/agent/internal/goaldefinition`)
  - `src/agent/internal/goaldefinition` - blir tunn wrapper eller tas bort
  - `src/nuxt/app/components/DefinitionPanel.vue` (ny)
  - `src/nuxt/app/composables/useGoalDefinition.ts` (ny)
  - `src/nuxt/app/pages/goal/[id].vue` - `Space+f`, tangent `Space+c` ->
    kriterier, varning om saknade fält
  - `src/nuxt/app/components/InboxGoalsList.vue` - "Bearbeta" -> panelen
  - `src/nuxt/app/components/HelpModal.vue` - nya tangenter
- **Koppling till `add-goal-verification-and-phase-transitions`:** den changeen
  föreslår strukturerade verifieringsplaner (`goal_verifications`) och en
  fasmaskin där `define -> plan` kräver "en godkänd agentdefinition". Den här
  changeen ger samma definition - men som ett godkännande i gränssnittet, och
  fasmaskinen ska inte kräva den. Det här är **inte** blockerande: definitionen
  ska vara rådgivande, enligt beslutet nedan.
- Inga brytande ändringar av befintlig funktionalitet.

## Beslut

- **Definitionen är rådgivande.** `define -> plan` går alltid att göra. En
  varning visas i stället för ett 409.
- **Definitionen är ett förslag per session**, inte ett godkännandeobjekt i
  databasen. Den som vill ha historik över föreslagna definitioner lägger in
  `goal_definition_applies` - den loggar *vad som applicerades*, inte vad som
  föreslogs.
- **Synkront i Plan API:t.** Att definiera ett mål är en snabb mänsklig
  handling och ska inte vänta på en agentkörning.
- **Fältvis godkännande.** Ett godkännande av fält A skriver inte fält B.
- **Kriteriet är fritt.** Ett mål utan verifieringskriterium är inte fel.
