---
title: Mål kan ha en lång beskrivning, redigerad med tangenten t
type: note
tags: [feature, goals, go, nuxt, db, projekt:plan]
---

VARFÖR: titeln på ett mål är kort (VARCHAR(64)) och räcker inte när man vill skriva
ut vad som faktiskt ska uppnås. Det fanns ingen plats för längre text.

HUR: ny fritextkolumn "description" i goals (kanoniskt schema:
environments/base/db/schema.yaml - appliceras av Atlas Operator). API:t läser
och skriver den via PATCH /api/v1/goals/:id (UpdateGoalRequest.Description).
Frontend: tangenten 't' på /goals öppnar en textarea under det markerade målet
(Ctrl+Enter spara, Esc avbryt), på målsidan (/goal/:id) öppnas samma editor för
det markerade målet eller undermålet. Listan visar en förkortad förhandsvisning
(utils/goalDescription.ts: descriptionPreview).

FALLER: kolumnen är nullable (den fanns redan i dev-databasen på Neon som
"description text"), så alla SELECT kör COALESCE(description, '') - annars
kraschar scanningen på NULL. Kolumnen lämnas nullable i det kanoniska schemat
för att slå upp exakt samma DDL som redan finns i databaserna.

Verifierat: go test ./... ok, bun test ok (23 tester), bun run build ok, samt
e2e i Chromium (dev-login -> /goals -> 't' -> skriv -> Ctrl+Enter, och
/goal/:id -> 'k' -> 't' -> textan fylls med sparad text).
