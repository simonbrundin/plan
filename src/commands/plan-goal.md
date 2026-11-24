---
description: Hantera mål i goals.json
allowed-tools: bash,read,write
---

# Plan Goal

Hantera mål i Plan goal-store. Använd detta när användaren vill skapa, visa eller uppdatera mål.

## Användning

### Lista alla mål
Använd verktygen för att lista goals:
- `list_goals` - Lista alla mål
- `list_goal_statuses` - Visa tillgängliga statuser

### Skapa nytt mål
Använd `create_goal` med:
- `id`: Unikt ID (kebab-case)
- `title`: Målets titel
- `parent_id`: Förälder-mål-ID (valfritt)
- `status`: define, plan, execute, verify, completed

### Uppdatera mål
Använd `update_goal` med:
- `id`: Målets ID
- `status`: Ny status

## Exempel

**Skapa ett nytt mål:**
```
create_goal:
  id: "min-first-mål"
  title: "Första målet"
  parent_id: "root"
  status: "define"
```

**Lista alla mål:**
```
list_goals:
  parent_id: "root"
```

**Uppdatera status:**
```
update_goal:
  id: "min-first-mål"
  status: "completed"
```

## Goals.json

Mål sparas i `goals.json` i repot root. Detta är en temporär lösning tills Plan-applikationen kör stablt.
