---
title: Agentkörningar och godkännanden
description: Det kompletta flödet - från att du markerar ett mål till att agenten öppnar en Pull Request, och hur jobbet körs säkert i Kubernetes.
order: 5
icon: i-lucide-git-pull-request
updated: 2026-11-25
group: Agenten
---

# Agentkörningar och godkännanden

Det här är flödet när du markerar ett mål för agentarbete.

## Från mål till körning

```text
1. Du markerar målet för agent        PATCH /goals/:id  { "agent": true }
2. Du startar en körning              POST  /agent-runs { goal_id, phase, status: "queued" }
3. Jobbcontrollern ser kön            → skapar ett Kubernetes Job
4. Workern adopterar körningen        status: running
5. Agenten arbetar i workspace        define → plan → execute → verify
6. Human gate                        status: waiting_approval
7. Du beslutar                        approved / rejected
8. Koden skrivs, testas, PR skapas    status: succeeded
```

Steg 1 och 2 går att göra från gränssnittet eller med `curl`:

```bash
# Markera målet för agenten
curl -X PATCH http://localhost:4100/api/v1/goals/<id> \
  -H 'Content-Type: application/json' -d '{"agent":true}'

# Köa en körning
curl -X POST http://localhost:4100/api/v1/agent-runs \
  -H 'Content-Type: application/json' \
  -d '{"goal_id":<id>,"phase":"plan","status":"queued","agent_profile":"coding"}'
```

En körning som står i `queued` är en **uttrycklig begäran om att få
köras** - det gäller även om målet inte är markerat som agentmål.
Bekräftelsen får inte tappa vägen i något lager, för då står jobbet bara
och väntar.

## Körningen som Kubernetes Job

En AgentRun i `queued` blir ett Job i namespace `plan`:

```text
Plan-mål
   │
   ▼
AgentRun (status queued)          ← Plan äger körningen
   │
   ▼  jobbcontroller
Kubernetes Job (agent-worker-run-<id>)
   ├── init-container: klonar repot till /workspace/repo
   └── agent-worker: arbetar i en isolerad kopia, rapporterar till Plan
```

### Säkerhetsgränser

Jobbet är byggt för att inte kunna göra skada:

| Gräns | Värdet |
| --- | --- |
| ServiceAccount-token | Monteras inte (`automountServiceAccountToken: false`) |
| Egen identitet | `plan-agent-worker`, inga rättigheter i klustret |
| Resurser | CPU- och minnesgränser per profil |
| Tidsgräns | `ActiveDeadlineSeconds` (default 3600) |
| Städning | `TTLSecondsAfterFinished` |
| Omstart | `RestartPolicy: Never` |
| Kapacitet | `AGENT_JOB_MAX_CONCURRENT`, globalt (standard fyra) |
| Användare | non-root (UID/GID 65532) |
| Filsystem | read-only root filesystem, `tmpfs` för skrivbart |
| Capabilities | droppade, `runtime-default` seccomp |
| Hemligheter | från Secrets, aldrig från manifesten |

Jobb för olika mål kan köras parallellt, medan samma mål alltid
serialiseras.

### Image

Jobbens image läses ur agentens egen pod (`POD_NAME` + downward API), så
att jobben alltid kör exakt den kod som är utbyggd - inte en äldre tagg
som råkar ligga kvar i klustret.

### Git-hemligheten

`GIT_AUTH_HEADER` ska vara en komplett header, till exempel
`Authorization: Basic <base64(x-access-token:TOKEN)>`. Hemligheten
skapas utanför git. Agent-imagen innehåller även `gh`; efter en lyckad
verifiering skapar codern en idempotent Pull Request från
agent-branchen. Den befintliga tokenen i `GIT_AUTH_HEADER` används
tillfälligt som `GH_TOKEN` för PR-anropet, utan att skrivas till
git-konfiguration eller loggar.

Om PR:n redan finns för samma branch återanvänds dess URL. Om
PR-skapandet misslyckas markeras artifact-eventet som failed, men den
verifierade branchen förblir resultatet - en människa kan skapa PR:n
manuellt.

## När något går fel

| Situation | Vad agenten gör |
| --- | --- |
| Jobbet för en `running`-körning försvinner | Körningen markeras `failed` med ett `job_missing`-event - **inte** ett nytt jobb med okänd workspacehistorik |
| Jobbet överstiger `ActiveDeadlineSeconds` | Körningen rapporteras som misslyckad |
| Körningen står i `waiting_approval` när jobbet städas bort | Lämnas orörd - human gate kan fortfarande behöva slutföras |

Regeln om att inte starta om ett försvunnet jobb är viktig: en ny
körning i samma workspace kan ha en historik som inte längre stämmer,
och då vet ingen vad som faktiskt är gjort.

## Godkännanden

Godkännandet är det ställe där du fattar beslutet. Det visas på
`/agent`, och beslutet sparas i `agent_approvals`.

| Fält | Innehåll |
| --- | --- |
| `title` | Vad förslaget heter |
| `description` | Förslaget i klartext |
| `consequence` | Vad beslutet får till följd |
| `verification_plan` | Hur resultatet ska verifieras |
| `feedback` | Din återkoppling vid ett avslag |
| `status` | `pending`, `approved` eller `rejected` |
| `decided_by` / `decided_at` | Vem och när |

Ett avslag med feedback går tillbaka till agenten - feedbacken blir
input i nästa förslag.

## Händelser

Varje körning har en kronologisk händelsehistorik i `agent_run_events`.
Events har en typ (`type`), en fas och en JSONB-payload - till exempel
`job_missing`, `artifact` och usage-rapportering. Det är den historiken
som visas i gränssnittet under körningen.

## Kostnad per körning

`agent_run_usage` sparar tokenanvändning per körning: input-, output-
och cache-tokens, antal verktygsanrop och vilka verktyg som användes.
Plan räknar fram kostnaden ur prislistan, med prislistversionen
registrerad - så att kostnaden går att reproducera.

## Köra agenten lokalt

Workern kan köras mot ett enda mål, utan Kubernetes:

```bash
cd src
AGENT_GOAL_ID=449 go run ./agent/cmd/agent-worker
```

Utan `AGENT_REPO_PATH` gör workern bara definitioner - kodändringar
kräver ett repo att arbeta i.
