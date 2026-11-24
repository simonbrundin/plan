# Checkpoint-Gate Workflow

> Shopify Helix-inspirerat AI-kodande arbetsflöde med checkpoints, gates och orkestrerade sub-agenter.

## Översikt

Checkpoint-gate workflow är ett strukturerat sätt att låta AI-agenter bygga stora features genom att dela upp arbetet i små, verifierbara enheter. Varje checkpoint måste passera flera gates innan nästa checkpoint påbörjas.

```
┌─────────────────────────────────────────────────────────────┐
│                      WORKFLOW FLÖDE                          │
└─────────────────────────────────────────────────────────────┘

    Feature Request
          │
          ▼
┌─────────────────┐
│  Orchestrator   │
│  (main agent)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Checkpoint      │
│ Planner         │◄── Fresh context
│ (sub-agent)     │
└────────┬────────┘
         │
         ▼
   ┌─────────────┐
   │ Plan Review  │     👤 Human Approval
   │ (HTML View)  │
   └──────┬───────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────┐
│                    CHECKPOINT EXEKVERING                     │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ CP-1: Database Schema              [low complexity] │   │
│  │                                                     │   │
│  │ Tasks:                                             │   │
│  │  • Create users table                              │   │
│  │  • Define models                                   │   │
│  │  • Add migrations                                  │   │
│  │                                                     │   │
│  │ Gates:                                             │   │
│  │  ┌─────────┐ ┌─────────────┐                     │   │
│  │  │ Behavior│ │ Human Review │                     │   │
│  │  │  (auto) │ │   (manual)   │                     │   │
│  │  └────┬────┘ └──────┬──────┘                     │   │
│  │       │              │                             │   │
│  │       ▼              ▼                             │   │
│  │   go test        Test &                          │   │
│  │   golangci-lint   Approve                       │   │
│  │                                                     │   │
│  │  ✅ All gates passed → Next checkpoint           │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ CP-2: API Endpoints             [medium complexity]│   │
│  │                                                     │   │
│  │ Tasks:                                             │   │
│  │  • POST /register                                  │   │
│  │  • POST /login                                     │   │
│  │  • JWT middleware                                  │   │
│  │                                                     │   │
│  │ Gates:                                             │   │
│  │  ┌─────────┐ ┌─────────────┐ ┌────────────────┐  │   │
│  │  │ Behavior│ │ Adversarial │ │ Human Review   │  │   │
│  │  │  (auto) │ │  (critic+   │ │   (manual)     │  │   │
│  │  │         │ │   fixer)    │ │                │  │   │
│  │  └─────────┘ └─────────────┘ └────────────────┘  │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ... fortsätt tills alla checkpoints passerade ...         │
│                                                             │
└─────────────────────────────────────────────────────────────┘
          │
          ▼
       PR / MR
```

## Kärnprincip

> **"An attempt is allowed to be wrong. It is not allowed to ship until it isn't."**

Agenten får göra fel under implementation, men den **får inte gå vidare** till nästa checkpoint förrän alla gates har passerat.

## Checkpoints

Checkpoints är små arbetsenheter som:
- Bygger på varandra (ökar i komplexitet)
- Har tydliga, testbara uppgifter
- Tilldelas gates baserat på typ

### Komplexitetsnivåer

| Nivå | Exempel | Beskrivning |
|------|---------|-------------|
| `low` | Databas-schema | Infrastruktur, konfiguration |
| `medium` | API, affärslogik | Kärnfunktionalitet |
| `high` | Säkerhet, UI | Komplexa system |

### Exempel: "Add User Authentication"

```
CP-1: Database Schema [low]
├─ Tasks:
│  ├─ Create users table
│  ├─ Create sessions table
│  └─ Add indexes
└─ Gates: behavior, human

CP-2: API Endpoints [medium]
├─ Tasks:
│  ├─ POST /register
│  ├─ POST /login
│  ├─ POST /logout
│  └─ JWT middleware
└─ Gates: behavior, adversarial, human

CP-3: UI Login Form [medium]
├─ Tasks:
│  ├─ Email/password form
│  ├─ Form validation
│  └─ Error states
└─ Gates: behavior, ui, adversarial, human
```

## Gates

Gates är **exekverbara kvalitetskontroller** som blockerar agentens framfart.

### Gate 1: Behavior

**Syfte:** Verifierar att koden fungerar korrekt.

**Checks:**
- `go test ./...` - Enhetstester
- `golangci-lint run` - Kodstil
- `go vet ./...` - Statisk analys

**Blockerar:** Om något check failar.

### Gate 2: UI (valfri)

**Syfte:** Verifierar att UI matchar design.

**Implementation:**
- Jämför implementation med HTML-prototype
- Använder Claude/Gemini för spatial awareness
- Kollar spacing, sizing, position

**När:** Endast för checkpoints med UI-komponenter.

### Gate 3: Adversarial

**Syfte:** Hittar problem som agenten kan missa.

**Flöde:**
```
┌──────────────┐     ┌──────────────┐
│   Critic     │────►│    Fixer      │
│   Agent     │◄────│    Agent      │
└──────────────┘     └──────────────┘
       │                   │
       ▼                   ▼
   Hittar fel?         Fixar fel
       │                   │
       └───────┬───────────┘
               ▼
           Itera tills
         "inga fel hittade"
```

**Max iterationer:** 3

**Issue severities:**
- `critical` - Säkerhet, dataförlust
- `high` - Buggar, prestanda
- `medium` - Code smells
- `low` - Ignoreras ofta

### Gate 4: Human

**Syfte:** Manuell granskning.

**Blockerar:** Alltid (sista gate).

**Process:**
1. Agenten visar vad som implementerats
2. Människan testar funktionellt
3. Godkänn eller be om ändringar

## ShouldStop Hook ("Exit Code 2")

Den kritiska mekanismen som förhindrar agenten från att sluta för tidigt:

```go
// I AgentRuntime
if a.config.ShouldStop != nil && a.config.ShouldStop(turnCtx) {
    break  // Agenten ville sluta
}

// ShouldStop Hook returnerar false tills gates är passerade
func (o *Orchestrator) ShouldStopHook() func(*core.TurnContext) bool {
    return func(ctx *core.TurnContext) bool {
        for _, gateType := range checkpoint.Gates {
            if result.Status != GatePassed {
                return false  // "Fortsätt arbeta!"
            }
        }
        return true  // Gates klara, tillåt stop
    }
}
```

**Flöde:**
1. Agenten försöker avsluta (stopreason = end_turn)
2. Hook kollar om alla gates passerat
3. Om inte → returnerar `false` + feedback
4. Agenten fortsätter arbeta
5. Loopa tills alla gates passerat

## Arkitektur

### Mappstruktur

```
src/agent/pkg/core/
├── checkpoint/              # Data structures
│   ├── checkpoint.go       # Checkpoint, ComplexityLevel
│   └── plan.go            # CheckpointPlan, GateResult, Issue
├── gate/                   # Gate implementations
│   ├── gate.go            # Gate interface & registry
│   ├── behavior.go        # BehaviorGate
│   ├── ui.go             # UIGate
│   ├── adversarial.go     # AdversarialGate
│   └── human.go          # HumanGate
├── orchestrator/           # Orkestrering
│   ├── orchestrator.go   # Main orchestration logic
│   ├── state.go         # StateStore (InMemory/File)
│   ├── shouldstop.go    # ShouldStop hook
│   ├── planner.go      # Checkpoint planner (heuristic/LLM)
│   ├── viewer.go       # Console & HTML viewers
│   ├── cli.go         # CLI adapter
│   └── learnings.go    # Learnings manager
└── workflow/
    └── orchestrator_integration.go  # Global orchestrator
```

### API

**Base URL:** `http://localhost:8080/api/v1`

| Method | Endpoint | Beskrivning |
|--------|----------|-------------|
| POST | `/workflow` | Starta ny workflow |
| GET | `/workflow/plans` | Lista alla plans |
| GET | `/workflow/:id` | Hämta specifik plan |
| POST | `/workflow/:id/approve` | Godkänn plan |
| POST | `/workflow/:id/reject` | Avvisa plan |
| GET | `/workflow/current/checkpoint` | Nuvarande checkpoint |
| GET | `/workflow/current/progress` | Gate progress |
| POST | `/workflow/decision` | Submit human decision |

### API Exempel

**Starta workflow:**
```bash
curl -X POST http://localhost:8080/api/v1/workflow \
  -H "Content-Type: application/json" \
  -d '{
    "feature": "Add user authentication",
    "description": "Login, register, JWT sessions"
  }'
```

**Response:**
```json
{
  "plan": {
    "id": "abc-123",
    "feature": "Add user authentication",
    "description": "Login, register, JWT sessions",
    "checkpoints": [
      {
        "id": "cp-1",
        "name": "Database Schema",
        "complexity": "low",
        "gates": ["behavior", "human"],
        "tasks": ["Create users table", "Create sessions table"]
      }
    ],
    "status": "pending"
  },
  "message": "Plan created. Review and approve..."
}
```

**Godkänn plan:**
```bash
curl -X POST http://localhost:8080/api/v1/workflow/abc-123/approve \
  -H "Content-Type: application/json" \
  -d '{"approvedBy": "user@example.com"}'
```

## Användning

### 1. CLI

```bash
# Starta agenten
./agent

# Inuti chatten:
workflow start Add user authentication
checkpoint status
checkpoint progress
workflow approve
workflow reject
```

### 2. HTTP API

```bash
# Starta API-servern
./agent-api

# Skapa workflow
curl -X POST http://localhost:8080/api/v1/workflow \
  -d '{"feature": "Add notifications", "description": "..."}'

# Öppna Swagger UI
open http://localhost:8080/swagger/index.html
```

### 3. Frontend

Öppna `/checkpoints` i webbläsaren för en grafisk vy.

## Learnings

Feedback sparas automatiskt till `~/.config/agent/learnings.md`:

```markdown
# Learnings

## 2026-01-13: Checkpoint 3 (Permissions)

- **Issue**: Agent glömde admin-flaggan
- **Fix**: Lade till explicit check
- **Lärdom**: Testa alltid permission edge cases

## 2026-01-14: Checkpoint 5 (UI)

- **Issue**: Spacing var 8px istället för 12px
- **Fix**: Uppdaterade design spec
- **Lärdom**: Ha exakta pixel-värden i design.md
```

Alla nya agent-sessioner läser learnings före arbetet börjar.

## CLI Workflow

```
$ ./agent
💬 Chat started. Model: gpt-4o-mini (openai)
Type 'quit' or 'exit' to end.
Type 'workflow' to start a checkpoint-gate workflow.

You: workflow start Add user authentication

======================================================================
CHECKPOINT PLAN: Add user authentication
======================================================================

Description: Login, register, JWT sessions

Checkpoints (3):

  1. Database Schema [low]
     Set up database schema and models
     Tasks:
       • Create/update database schema
       • Define models
       • Add migrations if needed
     Gates: behavior, human

  2. API Endpoints [medium]
     Implement API endpoints
     Tasks:
       • Design API contracts
       • Implement handlers
       • Add input validation
       • Write API tests
     Gates: behavior, adversarial, human

  3. User Interface [medium]
     Implement user interface
     Tasks:
       • Create component structure
       • Implement styling
       • Add interactions
       • Test responsive design
     Gates: behavior, ui, adversarial, human

----------------------------------------------------------------------
Review and approve this plan before execution.
======================================================================

Do you approve this plan? (y/n): y
Plan approved. Workflow ready to execute.

You: quit
👋 Goodbye!
```

## Skills

### orchestrator
Huvudsaklig orkestrerings-skill.

### checkpoint-planner
Skapar checkpoint-planer från features.

### adversarial-loop
Koordinerar critic + fixer loop.

## Konfiguration

### Miljövariabler

| Variabel | Beskrivning |
|----------|-------------|
| `ORCHESTRATOR_MAX_ADVERSARIAL_ITERATIONS` | Max critic+fixer iterationer (default: 3) |
| `ORCHESTRATOR_GATE_TIMEOUT` | Timeout för gates (default: 30m) |
| `LEARNINGS_FILE` | Path till learnings-fil |

### Exempel

```bash
export ORCHESTRATOR_MAX_ADVERSARIAL_ITERATIONS=5
export ORCHESTRATOR_GATE_TIMEOUT=1h
export LEARNINGS_FILE=/home/user/.my-learnings.md
./agent
```

## Integration med GitOps

Workflowet passar naturligt med GitOps-arbetssätt:

```
Feature Request (GitHub Issue)
         │
         ▼
   Orchestrator startar
         │
         ▼
   Agent arbetar i isolerad branch
         │
         ▼
   Checkpoint-plan → PR description
         │
         ▼
   Gates verifierar (CI/CD)
         │
         ▼
   Human review (PR review)
         │
         ▼
     Merge
```

## Felsökning

### Agenten stannar inte

Kontrollera att ShouldStop hook är konfigurerad:
```go
config.ShouldStop = orchestrator.ShouldStopHookFunc()
```

### Gates passerar inte

1. **Behavior gate failar:**
   - Kör `go test ./...` manuellt
   - Kör `golangci-lint run`

2. **Adversarial gate hittar för mycket:**
   - Justera max iterations
   - Ignorera low-severity issues

3. **Human gate blockerar:**
   - Använd `/workflow/decision` endpoint
   - Eller godkänn via frontend

### State förloras

Standard är in-memory state. För persistence:
- Använd `FileStateStore`
- Eller implementera databas-backed store

## Licens

Detta är en del av agents-projektet.
