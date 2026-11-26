---
title: Fyra paritetsfällor i agentdomänens port till Rust
type: note
tags: [rust, migrering, agent, paritet, bug, projekt:plan]
---

Agentdomänen i Plan-API:t portades till Rust (körningar, händelser, godkännanden, budgetar, användning, API-nycklar). Live-pariteten mot Go avslöjade fyra fallgropar som inte syns i källan.

**1. `QueryBuilder::separated` dubblerar kommat när text och bindning blandas.** `separated(", ").push("status = ").push_bind(v)` ger `status = , $1` - `push` sätter separatorn, och `push_bind` sätter den igen eftersom flaggan fortfarande är satt. Använd repotets etablerade mönster i stället: manuell komma via `if antal > 0 { byggare.push(",") }` och `byggare.push(text).push_bind(värde)`. Samma fel uppstod för `updated_at = NOW()` - det är ett eget fält och behöver sin komma.

**2. `RETURNING` kan inte använda tabellprefixet.** En kolumnlista som `u.id, u.agent_id` fungerar i `SELECT` men ger syntaxfel i `RETURNING`. Håll två listor: en med prefix för läsning, en utan för `RETURNING`.

**3. `api_keys.user_id` är INTEGER, och `row.get` panikar vid typmismatch.** Kod som läste `user_id` som `i64` fick `ColumnDecode ... INT8 is not compatible with INT4` - i en `unwrap`, alltså en panik som dödade worker-tråden. Läs som `i32` och konvertera, och använd `try_get` i stället för `get` så att en typmismatch ger ett fel i stället för en panik. Samma INTEGER-fall gäller `agent_runs.id`, `agent_runs.goal_id`, `agent_approvals.*` och `agent_run_usage.*` - deras id:n är INTEGER medan `agent_run_events.id` och `agent_approvals.id` är BIGINT.

**4. Gin svarar 404 innan authen körs.** I Go sitter auth-middlevaren på gruppen `/api/v1`, så en okänd väg ger `404 page not found` i ren text - aldrig 401. Gin saknar dessutom hantering av metodfel: `GET` på en route som bara har `PATCH` ger samma 404, inte 405. Rust har därför auth-lagret på API-routern (inte på hela routern) och en egen `fallback` + `method_not_allowed_fallback` som svarar med Gins text. Utan det hade Rust svarat 401 på okända vägar och 405 på fel metod - två avvikelser som syns direkt i paritetsskriptet.

**Vad som ändrades till följd av detta:** `transitions` i livscykeln har `null` för terminala tillstånd (Go lämnar en tom lista som null), en ogiltig API-nyckel svarar `{"error":"Invalid API key"}` utan extra `message` (Go skiljer på "nyckeln finns inte" och "databasen är inte uppkopplad"), och `/docs` skickar vidare till `/docs/` med 301 precis som Gin gör.
