## Context

Plan har redan en definitionsmotor: `src/agent/internal/definer` (LLM:en svarar
med ett JSON-objekt) och `src/agent/internal/goaldefinition` (policy: fyll
luckor, skriv aldrig över användarens värde, returnera en `Proposal` med
`Fields` från -> till). Resultatet hamnar i `agent_approvals` och en människa
beslutar via `POST /api/v1/agent-approvals/{id}/decision`.

Det som inte finns är vägen för en människa som vill definiera ett mål:

- Agentkörningen kräver `goals.agent = true`, `agent-worker`, Kubernetes-Job och
  polling. Definitionen kommer som en konsekvens av att agenten råkar köa
  upp målet - inte för att användaren bad om hjälp.
- `agent_approvals` är kopplat till `agent_run_id` (`NOT NULL`, `ON DELETE
  CASCADE`). Ett godkännande utan agentkörning finns alltså inte i dag.
- Godkännandet är allt-eller-inget: `Proposal.Values` appliceras i sin helhet.

Begränsningar som formatet måste respektera:

- `src/api` och `src/agent` ligger i **samma Go-modul** (`src/go.mod`,
  `module plan`), och `anthropic-sdk-go` finns redan i `require`-listan. En
  LLM-klient i Plan API:t är därför billig att lägga till.
- Agentens provider (`src/agent/internal/core/provider`) hanterar minimax och
  openai. Plan API:t behöver inte samma bredd - Anthropic-kompatibla
  minimax/openai går över samma endpoint-form.
- Källan till sanning för databasschemaet är
  `environments/base/db/schema.yaml` (ConfigMap), med
  `environments/pr/template/db/schema.sql` som PR-kopia. Migrations körs
  idempotent i `src/api/internal/database/db.go`.

## Goals / Non-Goals

- Goals:
  - En användare ska kunna definiera ett mål från inboxen eller målsidan utan
    agentflagga, agent-API:t eller väntan på en worker.
  - Förslaget ska fylla luckor, aldrig skriva över användarens värden.
  - Användaren ska godkänna per fält, och bara de godkända fälten ska skrivas.
  - Ett mål ska kunna få ett verifieringskriterium som en del av definitionen.
  - `define -> plan` ska aldrig blockeras av definitionen.
- Non-Goals:
  - Ingen agent som arbetar på målet (det är `add-agent-assisted-goals`).
  - Ingen fasmaskin med villkor (det är
    `add-goal-verification-and-phase-transitions`).
  - Ingen strukturerad verifieringsplan med typ/styrka och bevis - här är
    kriteriet en enkel fritextlista. Strukturen kan komma i den andra changeen.
  - Ingen historik över *föreslagna* (ej applicerade) definitioner.
  - Ingen batch-definering av hela inboxen.
  - Ingen "mig en vana"-riktning (föreslå nya mål) - det är en egen change.

## Decisions

- Decision: LLM:anropet sker i **Plan API:t** (`src/api/internal/llm`), synkront
  i endpointen.
  - Rationale: definitionen är en snabb, mänsklig handling. Agentmaskineriet
    behåller sitt jobb (planera/utföra kod), men ska inte stå i vägen för att
    definiera.
  - Alternativ som övervägdes: (a) anropa agent-tjänsten - blir asynkront
    igen, kräver agentflagga och jobb; (b) köa via `agent-worker` - samma
    problem plus extra led. Båda förkastas för att de gör definitionen långsam
    och svårförståelig.
- Decision: Förslaget lever i **klienten** tills användaren godkänner.
  - Rationale: "godkänna allting" blir bokstavligt - inget skrivs förrän
    godkännande skickas, och bara de godkända värdena skickas.
  - Alternativ som övervägdes: spara förslaget i `agent_approvals` - kräver
    `agent_run_id` och binder definitionen till agentmaskineriet.
- Decision: Appliceringen loggas i **`goal_definition_applies`** (mål, fält,
  från, till, tidpunkt).
  - Rationale: spårbarhet utan att föreslaget behöver överleva. Tabellen
    dokumenterar *vad som applicerades*, av vem och när.
- Decision: Policyn ("fyll luckor, skriv aldrig över") ligger i ett **gemensamt
  paket** `src/internal/goaldefinition`, som både Plan API:t och agenten
  använder.
  - Rationale: två definitioner av samma policy är två som glider isär. Flytten
    är liten - paketet har inga beroenden till Plan eller agent.
  - Alternativ som övervägdes: duplicera policyn - billigare nu, dyrt senare.
- Decision: Verifieringskriteriet är **fritext** på målet (`goals.criteria`).
  - Rationale: det räcker som acceptanskriterium ("testet är grönt") och
    binder ihop faserna. Den strukturerade varianten med typ/styrka och bevis
    finns redan som förslag i
    `add-goal-verification-and-phase-transitions`.
  - Alternativ som övervägdes: kräva kriterium för att lämna `define` -
    förkastat, definitionen ska vara rådgivande.
- Decision: Definitionen **blockerar inte** fasbyten. En varning visas i stället.
  - Rationale: definitionen kan göras i efterhand, och en måls definition kan
    förbättras senare utan att låsa något.
- Decision: Tangenten för att öppna panelen är **`Space+f`** (leader + f).
  - Rationale: `d` är upptaget ("toggle slutfört" på målsidan, och `Space+d`
    sätter startdatum/deadline). `f` är ledig, och leader-läget är i appen
    avsedd för sekundära åtgärder (`Space+d`, `Space+t`, `Space+v`).
  - Alternativ som övervägdes: plain `g` - `g` är globalt prefix för
    navigation (`g i`, `g g`, `g n`) och blir förvirrande; `Space+df` - för
    långt.

## Risks / Trade-offs

- LLM:en i Plan API:t blir en ny beroende i ett API som annars är
  deterministiskt. Mitigation: LLM:en når bara definitionens `propose`, aldrig
  `apply`; `apply` är ren validering och skriver de värden klienten skickar.
  Saknas API-nyckel ska `propose` svara med ett tydligt fel och panelen
  fungera utan förslag (luckorna visas, användaren fyller i dem själv).
- Ett förslag kan bli fel. Mitigation: fältvis godkännande, antaganden visas,
  och inget skrivs utan godkännande.
- Två klienter (agentens och Plan API:ts) kan producera olika format.
  Mitigation: gemensamt paket för policy och format.
- Kostnad: varje förslag kostar ett LLM-anrop. Mitigation: bara vid uttrycklig
  begäran, aldrig i bakgrunden.
- `criteria` som fritext kan bli inkonsekvent mellan mål. Mitigation:
  fritextformatet är medvetet löst nu; den strukturerade varianten finns i den
  andra changeen.

## Migration Plan

1. `environments/base/db/schema.yaml`: lägg till `criteria` på `goals` och
   tabellen `goal_definition_applies`.
2. `environments/pr/template/db/schema.sql`: samma ändring.
3. `src/api/internal/database/db.go`: idempotenta `ALTER TABLE ... ADD COLUMN
   IF NOT EXISTS` och `CREATE TABLE IF NOT EXISTS` - samma mönster som
   befintliga migreringar.
4. Ny endpoint fungerar så snart kolumnen finns; `criteria` är nullable och
   gamla mål läses som tomma.

Rollback: revertera committen. Kolumnen och tabellen kan ligga kvar - de
skadar inte.

## Open Questions

- Ska `goal_definition_applies` även lagra de *avvisade* fälten, så att det
  går att se vad agenten föreslog men du tackade nej till? (Just nu: nej.)
- Ska kriteriet kunna vara flera punkter i en lista, eller ett fritt fält med
  markdown? (Just nu: fritt fält, listor skrivs som rader.)
- Ska definitionen kunna köras utan LLM (rent lokalt, deterministisk
  luckinventering)? (Just nu: ja - panelen visar luckorna även utan förslag.)
