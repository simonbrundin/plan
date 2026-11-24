---
name: checkpoint-planner
description: |
  Tar en stor feature och delar upp den i sekventiella checkpoints
  med ökande komplexitet.
  
  Utlösare: "plan checkpoints", "dela upp feature", "skapa checkpoints"
---

# Checkpoint Planner Skill

Skapar checkpoint-planer för features.

## Input

En feature-beskrivning, t.ex.:

```
Implementera ett användarautentiseringssystem

Krav:
- Login med email/lösenord
- Registrering med email-verifiering
- Återställning av lösenord
- JWT-baserade sessions
- Role-based access control
```

## Output

```json
{
  "feature": "Användarautentisering",
  "description": "...",
  "checkpoints": [
    {
      "id": "cp-1",
      "name": "Databas Schema",
      "complexity": "low",
      "gates": ["behavior"],
      "tasks": [
        "Skapa users-tabell",
        "Skapa sessions-tabell",
        "Lägg till index"
      ]
    },
    {
      "id": "cp-2",
      "name": "API Endpoints",
      "complexity": "medium",
      "gates": ["behavior", "adversarial"],
      "tasks": [
        "POST /register",
        "POST /login",
        "POST /logout",
        "POST /forgot-password"
      ]
    }
  ]
}
```

## Checkpoint-regler

1. **Ordning** - Checkpoints ordnas med ökande komplexitet
2. **Beroenden** - Varje checkpoint bygger på den föregående
3. **Gates** - Tilldelas baserat på checkpoint-typ:
   - `behavior` - Alltid
   - `adversarial` - Kod-intensive checkpoints
   - `ui` - Endast för UI/frontend-arbete
   - `human` - Alla checkpoints

## Komplexitetsnivåer

- **low** - Databas-schema, konfiguration, infrastructure
- **medium** - API:er, affärslogik, UI-komponenter
- **high** - Säkerhet, integrationer, komplexa flöden

## Användning

```bash
# Internellt anrop från orchestrator
checkpoint_planner.create_plan(
  feature="Användarautentisering",
  description="Fullständigt auth-system med email och RBAC"
)
```

## Limitationer

- Planner genererar en baseline - människan granskar och justerar
- Test-planering görs separat (av TDD-planner)
