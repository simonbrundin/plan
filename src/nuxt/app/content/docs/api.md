---
title: API och autentisering
description: Plan API:t, agent-API:t, hur sessioner och API-nycklar fungerar, och vilka scopes agenten får.
order: 6
icon: i-lucide-plug
updated: 2026-11-25
group: Utvecklare
---

# API och autentisering

Det finns två API:er. Plan API:t äger domändata och autentisering, och
agent-API:t hanterar chatt, verktyg och jobb. Gränssnittet pratar med
båda, men alltid via Nuxt:s server - aldrig direkt från webbläsaren
för agent-API:t.

## Plan API

Go + Gin, bas `/api/v1`, port 8080. Ligger i `src/api`.

### Autentisering (utan autentisering)

| Metod | Sökväg | Beskrivning |
| --- | --- | --- |
| GET | `/api/v1/auth/login` | Starta OAuth-inloggning mot Zitadel |
| GET | `/api/v1/auth/callback` | OAuth-callback - skapar sessionen |
| POST | `/api/v1/auth/dev-login` | Dev-inloggning (bara när `devAuthEnabled`) |
| GET | `/api/v1/auth/session` | Vilken session är giltig |
| POST | `/api/v1/auth/logout` | Logga ut |

### Mål

| Metod | Sökväg |
| --- | --- |
| GET | `/api/v1/goals` |
| GET | `/api/v1/goals/prioritized` |
| GET | `/api/v1/goals/graph` |
| POST | `/api/v1/goals` |
| GET | `/api/v1/goals/:id` |
| PATCH | `/api/v1/goals/:id` |
| DELETE | `/api/v1/goals/:id` |
| POST | `/api/v1/goals/:id/completions` |
| DELETE | `/api/v1/goals/:id/completions` |
| PATCH | `/api/v1/goals/reorder` |
| PATCH | `/api/v1/goals/priority/reorder` |

### Status, relationer och beroenden

| Metod | Sökväg |
| --- | --- |
| GET | `/api/v1/statuses` |
| GET | `/api/v1/goals/:id/status` |
| PATCH | `/api/v1/goals/:id/status` |
| GET | `/api/v1/goals/:id/status/history` |
| POST | `/api/v1/goals/relations` |
| PATCH | `/api/v1/goals/relations` |
| PATCH | `/api/v1/goals/relations/weights` |
| DELETE | `/api/v1/goals/relations` |
| PATCH | `/api/v1/goals/relations/reorder` |
| POST | `/api/v1/goals/dependencies` |
| DELETE | `/api/v1/goals/dependencies` |

### Kalender, platser och taggar

| Metod | Sökväg |
| --- | --- |
| GET/POST | `/api/v1/calendars` |
| PATCH/DELETE | `/api/v1/calendars/:id` |
| GET/POST | `/api/v1/calendar-events` |
| GET/PATCH/DELETE | `/api/v1/calendar-events/:id` |
| POST | `/api/v1/calendar-events/:id/occurrences` |
| POST | `/api/v1/calendar-events/:id/split` |
| GET/POST | `/api/v1/places` |
| GET | `/api/v1/places/nearby` |
| PATCH/DELETE | `/api/v1/places/:id` |
| GET/PUT | `/api/v1/goals/:id/places` |
| GET/PUT | `/api/v1/calendar-events/:id/places` |
| GET | `/api/v1/tags` |
| GET/PUT | `/api/v1/goals/:id/tags` |
| DELETE | `/api/v1/goals/:id/tags/:name` |
| GET/PUT | `/api/v1/me/current-place` |

### Agentkörningar

| Metod | Sökväg |
| --- | --- |
| GET | `/api/v1/agent-run-lifecycle` |
| GET/POST | `/api/v1/agent-runs` |
| PATCH | `/api/v1/agent-runs/:id` |
| POST | `/api/v1/agent-runs/:id/cancel` |
| GET/POST | `/api/v1/agent-runs/:id/events` |
| GET/POST | `/api/v1/agent-runs/:id/approvals` |
| GET | `/api/v1/agent-approvals` |
| POST | `/api/v1/agent-approvals/:id/decision` |
| GET/POST | `/api/v1/agent-runs/:id/usage` |
| GET | `/api/v1/agent-usage` |
| GET/POST | `/api/v1/agent-budgets` |
| PATCH/DELETE | `/api/v1/agent-budgets/:id` |
| GET | `/api/v1/agent-budgets/:id/status` |

### Övrigt

| Metod | Sökväg | Beskrivning |
| --- | --- | --- |
| GET | `/health` | Health check, utan auth |
| GET | `/docs/*` | Scalar UI med OpenAPI-specen |
| GET/POST | `/api/v1/api-keys` | Skapa och lista API-nycklar |
| DELETE | `/api/v1/api-keys` | Ta bort nyckel |

## Agent-API

Go + Gin, bas `/api/v1`, port 8080, egen deployment (`agent-api`) och
egen databas (`plan-agent-db`). Ligger i `src/agent`.

| Metod | Sökväg | Beskrivning |
| --- | --- | --- |
| GET | `/api/v1/health` | Health check |
| GET | `/api/v1/status` | Status |
| POST | `/api/v1/workflow` | Starta en workflow |
| GET | `/api/v1/workflow/plans` | Lista planer |
| GET | `/api/v1/workflow/:id` | En plan |
| POST | `/api/v1/workflow/:id/approve` | Godkänn plan |
| POST | `/api/v1/workflow/:id/reject` | Avslå plan |
| GET | `/api/v1/workflow/current/checkpoint` | Aktuell checkpoint |
| GET | `/api/v1/workflow/current/progress` | Förlopp |
| POST | `/api/v1/workflow/decision` | Mänskligt beslut |
| GET | `/api/v1/goals` | Mål som agenten får arbeta med |
| POST | `/api/v1/goals/start-agent` | Starta agent för ett mål |
| GET/POST | `/api/v1/conversations` | Samtal |
| GET | `/api/v1/conversations/:id/messages` | Meddelanden |
| POST | `/api/v1/conversations/:id/messages` | Skicka meddelande |
| GET | `/api/v1/conversations/:id/stream` | Strömning av svar |
| GET | `/api/v1/approvals` | Väntande godkännanden |
| POST | `/api/v1/approvals/:goal_id/decision` | Beslut |
| GET | `/api/v1/api-keys` | Modellnycklar |
| GET | `/api/v1/ws/chat` | WebSocket-chatt |

## Autentisering

### Sessioner

**Authen ligger i Go API:t**, inte i Nuxt. Go skapar sessionen, sätter
två cookies och validerar sessionen från cookie *eller* `Authorization:
Bearer`:

| Cookie | Innehåll |
| --- | --- |
| `plan_session` | httpOnly - sessionstokenet |
| `plan_user` | Läsbar hint - till exempel användarens namn |

Sessionstoken får inte cirkula i URL:er - det var det gamla
callback-flödet.

### Flödet i Nuxt

Nuxt är en tunn proxy (`server/utils/goApi.ts`): den vidarebefordrar
cookien och kopierar tillbaka `Set-Cookie` + `Location` från Go.

```text
Webbläsare → /api/proxy/goals → Nuxt (forwardToGo) → Go API
     ↑                                                 │
     └────────── Set-Cookie + JSON-svar ──────────────┘
```

Agent-API:t anropas bara server till server, med `AGENT_API_KEY` - den
är aldrig i webbläsaren.

### API-nycklar

Programmatisk åtkomst sker med API-nycklar. Nyckeln lagras hashad
(`key_hash`), tillsammans med scope:

| Scope | Ger |
| --- | --- |
| `full` | Allt som en session kan göra |
| `agent` | Endast de anrop som står med `agent: true` i kontraktet |

Agentens nyckel har scope `agent` i produktion. Därför måste Plan
tillåta exakt de anropen agenten gör - annars blir det 403 i prod,
medan dev (där nyckeln har scope `full`) inte visar felet. Scope-spärren
implementeras i `internal/middleware/agent_scope.go` och testas mot
`src/contracts/plan-agent.yaml`.

### Miljövariabler

| Variabel | Vart | Beskrivning |
| --- | --- | --- |
| `NUXT_PUBLIC_GO_API_URL` | Nuxt | Publik URL till Plan API:t |
| `NUXT_GO_API_INTERNAL_URL` | Nuxt | Intern URL ( används i klustret) |
| `NUXT_AGENT_API_INTERNAL_URL` | Nuxt | Intern URL till agent-API:t |
| `NUXT_AGENT_API_KEY` | Nuxt | Nyckel mot agent-API:t |
| `DATABASE_URL` | Go | PostgreSQL-anslutning |
| `PLAN_API_URL` | Agent | Plan API:t |
| `PLAN_API_KEY` | Agent | Agentens nyckel mot Plan |

## OpenAPI

Plan API:t dokumenterar sig själv: `swagger.json` genereras ur
annotations i handlarna och serveras med Scalar UI på `/docs/`.
`src/api/internal/contracts` innehåller kontraktet som testas mot
agentens klient.
