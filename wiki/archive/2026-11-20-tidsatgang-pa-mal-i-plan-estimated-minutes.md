---
title: Tidsåtgång på mål i Plan (estimated_minutes)
type: note
tags: [feature, plan, goals, schema, frontend, backend, projekt:simon]
---

Funktion: alla mål i Plan kan ha en uppskattad tidsåtgång i minuter (nullable estimated_minutes).

Varför: uppskattningen ger ett sätt att se om en uppgift är stor eller liten utan att öppna den - och den behöver vara per mål (inte per uppgift) eftersom målen i Plan också är 'att göra'-poster.

Design:
- Enheten är minuter i databasen (estimated_minutes INTEGER, nullable). Formateringen (1 h 30 min) sker i presentationen, inte i lagringen - annars kan man inte summera och jämföra.
- Ett eget inmatningsfält (EstimateBadge.vue) används på målsidan, grundmålssidan, i listan och på prioriteringssidan. Tangenten 'm' öppnar/uppdaterar uppskattningen på de tre tangentbordsstyrda sidorna.
- Inmatningen accepterar 90, 1h, 1h30, 1,5h och 45m (utils/estimate.ts), och negativa värden avvisas av API:t (HTTP 400, 'tidsåtgången måste vara ett positivt antal minuter').
- PATCH /api/v1/goals/:id har en NullableInt för estimated_minutes: man kan sätta, cleara (null) eller lämna fältet orört.
- Ändringen går att ångra med Z (undo-stapeln), precis som övriga måländringar.

Verifiering: go test ./... (inkl. integrationstest mot dev-Neon), bun run build, 167 Nuxt-enhetstester, samt manuellt mot en lokalt startad API: PATCH sätter/clearaar, negativt ger 400, och /goals, /goals/graph, /goals/prioritized returnerar estimated_minutes.

Fynd under implementationen:
- environments/dev/db/neon.env kan inte sourcingas (variablerna blir tomma) eftersom URL:en innehåller '&'. Använd grep/cut istället: URL=$(grep '^DATABASE_URL=' environments/dev/db/neon.env | cut -d= -f2-).
- Kolumnen estimated_minutes fanns redan i dev-databasen (TOM, 160 rader) trots att schema.yaml inte hade den - den verkar ha lagts in manuellt tidigare. Schema.yaml i environments/base/db är nu i synk.
- Plan's API-nyckelautentisering fungerar inte (se notes/agent-api-crashade-i-tilt-...). I dev fungerar 'Authorization: Bearer user_1'.
