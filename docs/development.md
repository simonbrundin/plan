# Utveckling i dev

Plan och agenten bor i samma repository och deployas från **en** Tilt.
Två Tilt-instanser som deployar samma manifests fightar om
k8s-objekten - därför finns bara en.

## Förutsättningar

- Docker och Kubernetes (Kind via `devenv/dev.sh`)
- Tilt
- `kind`-klustret `kind-dev` med namespace `plan`
- En Neon-databas för Plan (`environments/dev/db/neon.env`, se
  `environments/dev/db/README.md`)

## Starta

```bash
cd environments/dev
tilt up
```

Tilt bygger tre images:

| Image | Byggcontext | Vad det är |
| --- | --- | --- |
| `plan-nuxt` | `src/nuxt` | Plan-gränssnittet |
| `plan-go-api` | `src/api` | Plan API - äger mål, körningar, godkännanden |
| `plan-agent` | repo-roten | Agent-API:t och workern - utför arbetet |

Portar i dev:

| Tjänst | Port |
| --- | --- |
| Plan (Nuxt) | http://localhost:3100 |
| Plan API | http://localhost:4100/api/v1 |
| Agent-API | http://localhost:8080 (port-forward) |

## Autonoma körningar

Körningar sker som Kubernetes Jobs. Jobbcontrollern i agent-API:t skapar ett
jobb per köad AgentRun:

```bash
# Lägg upp ett mål och markera det för agenten
curl -X PATCH http://localhost:4100/api/v1/goals/<id> \
  -H 'Content-Type: application/json' -d '{"agent":true}'

# Köa en körning
curl -X POST http://localhost:4100/api/v1/agent-runs \
  -H 'Content-Type: application/json' \
  -d '{"goal_id":<id>,"phase":"plan","status":"queued","agent_profile":"coding"}'
```

Jobbet klonar repot, föreslår en ändring, kör sitt eget test och skickar
förslaget till human gate. Beslutet tas i Plan - se
[`agent-platform.md`](agent-platform.md).

## Agentens databas

Agenten har en egen databas (`plan-agent-db`) så att agentens samtal inte
blandas med Plan. Hemligheter (modellnyckel, git-nyckel, Plan-nyckel) ligger
i Kubernetes Secrets och skrivs aldrig i git.

## Wiki

Wikin ligger i `wiki/` i det här repot och delas av alla andra repor.
Verktyget byggs med:

```bash
cd src/agent && go build -o ~/.local/bin/wikictl ./cmd/wikictl
```
