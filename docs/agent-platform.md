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

I prod är agenten inte utplockad ännu: `environments/prod/patches` tar bort
agentens deployment och databas tills hemligheterna (`plan-agent-api-key`,
`plan-agent-db-app`, `plan-agent-llm`, `plan-agent-git`) finns krypterade.

Det som återstår innan den gamla agents-repon kan arkiveras är:

- CI och releaseflöden för båda deployables (agent-go.yml tillkom, men
  image-publisering saknas)
- hantera flera samtidiga mål i jobbcontrollern
