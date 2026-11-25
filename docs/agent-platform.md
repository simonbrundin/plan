# Agentplattform

Plan är kontrollplanet och agenttjänsten är exekveringsmotorn. Båda ligger i
samma repository men byggs och körs som separata deployables.

```text
Plan API/Web
  ├── mål
  ├── AgentRuns
  ├── godkännanden
  ├── audit och usage
  └── proxy för agentchatt
          │
          ▼
Agent API (`src/agent`)
  ├── AgentRuntime
  ├── verktyg och gates
  ├── coding/life-profiler
  └── Kubernetes Job-körning
```

## Repository-layout

- `src/go/` — Plan API och kontrollplan.
- `src/nuxt/` — Plan-gränssnittet.
- `src/agent/` — agentmotorn, worker och agent-API:t som egen Go-modul.
- `src/commands/` — kommandon som kan laddas av agentmotorn.
- `wiki/` — gemensam kunskapswiki.

## Exekveringsmodell

1. Ett mål markeras för agentarbete i Plan.
2. Plan skapar eller visar en `AgentRun`.
3. Agenttjänsten hämtar målet och arbetar i en avgränsad profil.
4. Autonoma körningar ska köras som Kubernetes Jobs med egen identitet,
   explicit secrets, resursgräns och timeout.
5. Agenten rapporterar status, sammanfattning, artifacts och kostnad tillbaka
   till Plan.
6. Godkännanden ska visa mål, observationer, konsekvens och verifieringsplan —
   inte bara det tekniska kommandot.

## Autonoma körningar som Kubernetes Jobs

En AgentRun som står i `queued` är en uttrycklig begäran om att få köras.
Jobbcontrollern (`src/agent/internal/jobcontroller`) ser till att det finns
exakt ett Job per mål, och workern i jobbet adopterar den köade körningen
i stället för att skapa en egen.

Bekräftelsen är en uttrycklig begäran - den gäller även om målet inte är
markerat som agentmål. Workern hämtar då målet direkt. En begäran får inte
tappa vägen i något lager, för då står jobbet bara och väntar.

Om en running-körnings Job försvinner markeras körningen som `failed` med ett
`job_missing`-event i stället för att ett nytt jobb startas med okänd workspace-
historik. Jobb som överskrider `ActiveDeadlineSeconds` rapporteras också som
misslyckade. `waiting_approval` lämnas orörd när ett färdigt Job städats bort,
eftersom human gate fortfarande kan behöva slutföras.

```text
Plan-mål
   │
   ▼
AgentRun (status queued)          ← Plan äger körningen
   │
   ▼  jobbcontroller
Kubernetes Job (agent-worker-run-<id>)
   ├── init-container: klonar repot till /workspace/repo
   └── agentworker: arbetar i en isolerad kopia, rapporterar till Plan
```

Jobbet är byggt med säkerhetsgränser:

- ingen ServiceAccount-token monteras (`automountServiceAccountToken: false`)
- egen ServiceAccount (`plan-agent-worker`) utan Kubernetes-token
- CPU- och minnesgränser, deadline och TTL
- `RestartPolicy: Never`
- global kapacitet via `AGENT_JOB_MAX_CONCURRENT` (standard fyra); jobb för
  olika mål kan köras parallellt medan samma mål fortfarande serialiseras
- non-root (UID/GID 65532), read-only root filesystem, droppade capabilities,
  runtime-default seccomp
- API-, LLM- och Git-hemligheter kommer från Secrets, inte från manifest

Jobbens image läses ur agentens egen pod (`POD_NAME` + downward API), så att
jobben alltid kör exakt den kod som är utbyggd.

### Git-hemligheten

`GIT_AUTH_HEADER` ska vara en komplett header, till exempel
`Authorization: Basic <base64(x-access-token:TOKEN)>`. Hemligheten skapas
utanför git:

```bash
TOKEN=$(gh auth token)
HEADER="Authorization: Basic $(printf 'x-access-token:%s' "$TOKEN" | base64 -w0)"
kubectl --context kind-dev -n plan create secret generic plan-agent-git \
  --from-literal=GIT_AUTH_HEADER="$HEADER"
```

### Förslag verifieras innan human gate

Ett förslag testas med sitt eget testkommando innan det skickas till en
människa. Om testet inte går igenom får modellen först skriva om
testkommandot (innehållet behålls) och därefter skriva om hela förslaget.
Förslaget som människan ser är alltså självverifierat, och utfallet redovisas
i godkännandet.

## Agentprofiler

Profilerna är tekniska gränser, inte instruktioner till modellen. De ligger i
`src/agent/internal/profile` och tillämpas i tre lager:

| Gräns | coding | life |
| --- | --- | --- |
| Klona repot | ja | nej |
| Ändra filer | ja | nej |
| Git-hemlighet | ja | nej |
| Resursgräns | 1000m / 2Gi | 500m / 1Gi |
| Human gate | alltid | alltid |

Profilmoten följer med AgentRun (`agent_profile`) och sätts i jobbets miljö som
`AGENT_PROFILE`. En okänd profil ger kodprofilen - den strängaste - så ett
fel aldrig ger fler rättigheter än avsett.

Varje körning publicerar sin egen branch, `agent/goal-<mål>-run-<körning>`.
Då kan flera försök granskas side by side utan att krocka.

## Behörigheter

Behörigheterna är tekniska gränser på tre nivåer - ingen av dem ligger i
en prompt.

### Scopade API-nycklar

Plan API:t har två sorters nycklar. En full nyckel har samma rättigheter
som en inloggad användare. En agent-nyckel har `scope = agent` och får bara
arbeta med mål, agentkörningar och deras godkännanden:

| Anrop | Agent-nyckel |
| --- | --- |
| Läsa och skapa mål, sätta status | ja |
| Skapa, läsa och uppdatera körningar | ja |
| Läsa och skriva körningshändelser | ja |
| Skapa och läsa godkännanden | ja |
| **Besluta om ett godkännande** | **nej** |
| **Radera mål, körningar eller godkännanden** | **nej** |
| **Hantera API-nycklar, kalendrar, användare** | **nej** |

Human gate är alltså inte en fråga om vad agenten vill - agentens nyckel
kan inte fatta beslutet över huvudet på en människa. Nyckeln sätts i
`AGENT_API_KEY` och scopet finns i `api_keys.scope`.

### Agent-API:t kräver nyckel

Agent-API:t är bara tillgängligt för Plan. Nyckeln (`AGENT_API_KEY`) skickas
server-side från Nuxt, aldrig till webbläsaren. Utan nyckel loggar tjänsten
det tydligt vid start - det ska vara ett medvetet val, inte en olycka.

### Kubernetes

Jobben kör med `plan-agent-worker` som har `automountServiceAccountToken:
false` - de kan inte prata med Kubernetes-API:t alls. Agent-API:t har en egen
roll som bara får skapa, läsa och radera Jobb samt läsa poddar och loggar.

## Körningens livscykel

Varje körning börjar med `run_started` och slutar med `run_completed`.
Mellan dem ligger händelserna som beskriver vad som hände:

```text
run_started
  ├── definition_created        (målet specificerades)
  ├── approval_requested        (förslaget väntar på en människa)
  ├── approval_decided          (människan godkände eller avvisade)
  ├── phase_changed             (körningen bytte fas)
  ├── artifact                  (diff, branch, annat resultat)
  ├── verification              (testet kördes)
  └── run_completed             (körningen är klar)
```

En körning som avslutas av en profil som inte får röra kod avslutas
uttryckligen - den lämnas inte öppen bara för att profilen förbjuder ett
steg.

## AgentRun-statusar

En agentkörning har två dimensioner: en **fas** (vad agenten gör) och ett
**tillstånd** (hur det går). Båda är kanoniska och definieras i
`src/go/internal/models/agent_run.go` - klienter läser dem från
`GET /api/v1/agent-run-lifecycle` i stället för att hålla egna listor.

Faser, i ordning:

| Fas | Betyder |
| --- | --- |
| `define` | målet specificeras och bryts ned |
| `plan` | agenten planerar hur målet ska lösas |
| `execute` | agenten utför arbetet |
| `verify` | resultatet verifieras |
| `completed` | körningen är slut |

Tillstånd:

| Tillstånd | Betyder | Kan gå till |
| --- | --- | --- |
| `queued` | Plan har accepterat körningen, jobbet har inte startat | `running`, `waiting_approval`, `succeeded`, `failed`, `cancelled` |
| `running` | jobbet kör och agenten arbetar | `waiting_approval`, `succeeded`, `failed`, `cancelled` |
| `waiting_approval` | agenten väntar på en människa - till exempel ett godkännande | `queued`, `running`, `succeeded`, `failed`, `cancelled` |
| `succeeded` | körningen lyckades | terminalt |
| `failed` | körningen misslyckades | terminalt |
| `cancelled` | körningen avbröts | terminalt |

Reglerna är: en körning får alltid avbrytas, en terminal körning får inte
återupplivas, och fasen får aldrig gå bakåt.

**Namnen ändras inte utan migrering.** De ligger i databasen, i API-svar och i
gränssnittet - ett nytt namn utan migrering ger två sanningar. 2026-11-25 byttes
`blocked` till `waiting_approval` och `done` till `succeeded`; migreringen sker
idempotent i `src/go/internal/database/db.go` vid uppstart.

## Kontraktet mellan Plan och agenten

Plan äger agentkörningarna, men agenten och Plan pratar via HTTP - och de två
sidorna ligger i samma repo. Därför finns kontraktet som en fil:
`src/contracts/plan-agent.yaml`.

Filen beskriver de delar agenten använder: faser, tillstånd, terminala
tillstånd, tillstånd där en människa måste agera, triggrar, profiler och de
API-anrop agenten gör. Den är versionsstyrd - `version: 1` beskriver första
versionen, som samtidigt bytte `blocked` till `waiting_approval` och `done` till
`succeeded`.

Båda sidor testar mot samma fil:

| Sida | Test | Vad det fångar |
| --- | --- | --- |
| Plan | `src/go/internal/contracts` | att faser, tillstånd och övergångar i modellen stämmer med kontraktet, och att varje anrop i kontraktet finns i OpenAPI-specifikationen |
| Agent | `src/agent/internal/planclient` | att agentens statuskonstanter och sökvägar stämmer med kontraktet, och att kontraktets anrop finns i Plan |

Sökvägarna i agentens klient ligger i `src/agent/internal/planclient/paths.go`
som namngivna konstanter - annars går det inte att jämföra dem mot kontraktet.
I koden skrivs parametrar som `%d`, i kontraktet som `{id}`.

Om kontraktet ändras så att det kräver samtidiga ändringar på båda sidor, bumpas
versionen - då går det att se i historiken att kontraktet ändrades.

## Checkpoint-orchestratorn

Checkpoint-orchestratorn (`src/agent/internal/core/workflow`) driver det planerade
läget: ett mål bryts ned i checkpoints, och varje checkpoint passerar en gate
(beteende, UI, adversarial, human) innan nästa får köras.

**StateStoren bestämmer om läget överlever en omstart.** I dev och prod används
`orchestrator.NewCheckpointStore(pool)` som sparar mål, checkpoints och
gate-resultat i Postgres (`goals`, `checkpoints`, `checkpoint_results`). Finns ingen
databas faller agenten tillbaka på `NewInMemoryStateStore()`, och då är
checkpointplanen flyktig.

**Initieringsfel ska synas, inte bara loggas.** Utan orchestratorn registreras
inga workflow-endpoints - `/api/v1/workflow`, `/api/v1/workflow/:id/approve` och
gates saknas helt. `/api/v1/health` och `/api/v1/status` rapporterar därför
`orchestrator.initialized` och `orchestrator.error`, så att ett fel går att se
utan att läsa loggar.

Testerna i `internal/core/workflow` och `internal/api` håller båda delarna på
plats: att en egen StateStore behålls, att alla fyra gate-typerna är
registrerade, och att initieringsfelet syns i status och health.

## Bygg och driftsättning

| Deployable | Källa | Image | CI |
| --- | --- | --- | --- |
| Plan (Nuxt + Go API) | `src/nuxt`, `src/go` | `plan`, `go-api` | `ci.yaml`, `commit.nuxt.yaml`, `commit.api.yaml` |
| Agenten | `src/agent` | `plan-agent` | `commit.agent.yaml` |

Agentens CI bygger och publicerar `plan-agent` och föreslår en PR med den
nya imagen i `environments/base/agent-deployment.yaml`. Dev bygger samma
image lokalt via Tilt.

### Hemligheter i prod

Agentens hemligheter är inte i git. De skapas med:

```bash
cd environments/prod/secrets
SOPS_AGE_KEY_FILE=~/.config/sops/age/keys.txt \
  AGENT_PLAN_API_KEY=... AGENT_WORKER_KEY=... \
  AGENT_LLM_API_KEY=... AGENT_GIT_HEADER=... AGENT_DB_PASSWORD=... \
  ./create-agent-secrets.sh
```

Skriptet krypterar varje hemlighet med repoets publika age-nyckel. Hemligheterna
finns nu krypterade i `environments/prod/secrets/agent-*.enc.yaml` och agenten är
**inte längre borttagen** i prod-overlayson - `environments/prod/patches` innehåller
bara de patchar som fortfarande behövs. När klustret är igång är agenten redo att
deployas.

Kvar att göra i prod:

1. Kör `environments/prod/secrets/plan-api-key.sql` i Plan-databasen - den lägger in
   hashen av agentens API-nyckel (nyckeln finns krypterad, Plan sparar bara hashen).
2. Låt Flux synka `environments/prod` - agent-imagen hämtas från ghcr.io och
   jobbcontrollern börjar skapa jobb för köade körningar.

### Gates i CI

| Gate | Var | Vad den fångar |
| --- | --- | --- |
| `go vet` + `go test -race` | `agent-go.yml`, `commit.api.yaml` | fel i agenten och Plan-API:t |
| `golangci-lint` | `agent-go.yml` | linterfynd i agenten |
| `bun run test:unit` | `ci.yaml` | enhetstesterna i `src/nuxt/test/unit` (233 st) |
| `bun run test` (cucumber) | `ci.yaml` | feature-flöden i gränssnittet |
| `bun run typecheck` | `ci.yaml` | nya typfel - se nedan |

Typecheck-gaten är en baslinje-gate: gränssnittet har ärvd teknisk skuld (170 typfel
i `app/` och `server/`), så gaten jämför antalet fel per fil mot
`src/nuxt/typecheck-baseline.json` och faller bara på **nya** fel. Kör
`node scripts/typecheck.mjs --update` i `src/nuxt` för att skriva om baslinjen
efter en avsiktlig rättning.

## Kostnad och budget

Plan äger kostnaden. Agenten rapporterar bara tokenanvändningen - den
räknar inte själv, så att alla körningar prissätts likadant.

```text
agent                 Plan
  │                    │
  │ användning         │
  ├───────────────────►│  agent_run_usage
  │                    │  + kostnad ur prislista
  │                    │
  │ ny körning         │
  ├───────────────────►│  budgetkontroll + reservation
  │◄───────────────────┤  409 om gränsen är nådd
```

**Prislistan** versioneras (`src/go/internal/pricing`). Varje usage-rad sparar
prislistans version, valutan och om priset kom från listan eller från en
schablon. En okänd modell får aldrig kostnaden noll - då används en
dokumenterad schablonpris och raden markeras `fallback`.

**Budgetar** (`agent_budgets`) har en period, en hård gräns, en
varningsgräns och ett valfritt agent-id. Före varje ny körning kontrolleras
den förbrukade och den reserverade kostnaden i samma transaktion som
körningen skapas - därför kan inte två samtidiga körningar tillsammans
överskrida gränsen. En hård gräns ger `409` med `reason=limit_reached`, en
varningsgräns ger en varning i svaret men stoppar inte, och en okänd
modellpris ger `reason=unknown_price`.

Terminala körningar har alltid en usage-rad. Om agenten inte hann rapportera
markeras raden `incomplete`, så att det inte ser ut att kosta noll.

**Användningen måste rapporteras utan strömming.** MiniMax skickar
användningen i ett eget stycke i det strömmade svaret, och i det vanliga
svaret saknas den ibland helt. Därför frågar proposern efter användningen
(`StreamingOptions.Usage`), vilket gör att leverantören svarar utan
strömming - då kommer både tokenantal och kostnad med.

Översikten finns på `/agent-cost` i gränssnittet och via
`GET /api/v1/agent-usage` samt `GET /api/v1/agent-budgets`.

## Human gate i gränssnittet

Godkännandena ligger i Plan (`agent_approvals`), inte i agentens minne. UI:t
listar väntande förslag för målet, visar exakt de föreslagna ändringarna,
testkommandot och resultatet av agentens egen verifiering. Beslutet skickas
till Plan med människans session - agentens nyckel går inte igenom.

## Aktuell migrationsstatus

Agent-backendens kod är inflyttad till `src/agent` och bygger med
`src/agent/Dockerfile` från repository-roten. Plan har redan serverroutes och
UI för agentmål, körningar, konversationer och godkännanden.

Dev-miljön har nu manifest för en separat agentdatabas (`plan-agent-db`),
`agent-api` och Nuxt-proxy till agenten. Worker och scheduler är avstängda i
dev - autonoma körningar går via Kubernetes Jobs i stället.

Verifierat 2026-11-24 med mål #540: målet gick från Plan till köad AgentRun
(#63), Kubernetes Job (`agent-worker-run-63`), självverifierat förslag,
human gate (godkännande #50, beslutad via gränssnittet), applicerad ändring
i isolerad workspace, lyckad verifiering och publicerad branch
`agent/goal-540-run-63`. Huvudgrennen är orörd.

### Sammanslagningen av repon

Allt som behövdes från `agents`-repot finns i det här repot:

| Från agents | Hit | Läge |
| --- | --- | --- |
| `src/go` | `src/agent` | inflyttat och utvecklat vidare |
| `src/commands` | `src/commands` | identiskt |
| `wiki/` | `wiki/` | flyttat 2026-11-24, `~/wiki` pekar här |
| `src/go/cmd/wikictl` | `src/agent/cmd/wikictl` | byggs härifrån |
| `.github/workflows/go.yml` | `.github/workflows/agent-go.yml` | paths omdöpta |
| `environments/` | `environments/base/agent*` | dev och prod delar samma källa |
| `docs/` | `docs/` | se nedan |

Medvetet inte portade:

- **Agentens gamla Nuxt-gränssnitt** (checkpoints, kanban, readme, roadmap,
  repo-map) - Plan har egna sidor för mål, chatt och körningar. Koden finns
  kvar i `agents`-repot tills det arkiveras.
- **Det gamla flödet med filbaserade godkännanden och Slack** - godkännanden
  ligger nu i Plan, där en människa fattar beslutet.

I prod är agenten utplockad igen: hemligheterna (`plan-agent-api-key`,
`plan-agent-db-app`, `plan-agent-llm`, `plan-agent-git`) finns krypterade i
`environments/prod/secrets/`, och `environments/prod/patches` innehåller bara de
patchar som fortfarande behövs. Se "Hemligheter i prod" ovan för de två
stegen som återstår innan agenten kör i produktion.

Det som återstår för agents-repon:

- image-publisering och releaseflöden är på plats (`commit.agent.yaml` bygger
  `plan-agent` och föreslår PR med taggen), och CI-gates finns för agenten,
  Plan-API:t och gränssnittet
- hantera flera samtidiga mål i jobbcontrollern
