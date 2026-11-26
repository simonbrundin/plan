---
title: Human gate som teknisk gräns: scope på API-nyckeln
type: note
tags: [plan, agent, säkerhet, behörighet, human-gate, api-nyckel, projekt:simon]
---

I ~/repos/plan (Fas 4 i mål 497) görs human gate till en teknisk gräns i två steg.

1. Scopade API-nycklar i Plan (src/api). api_keys fick en scope-kolumn ('full' eller 'agent'). Agent-nyckeln har scope=agent och en allowlist (internal/middleware/agent_scope.go) som stoppar: beslut om godkännande, radering, och allt som hör till kalendrar, API-nycklar och användare. Agenten kan alltså inte godkänna sig själv - inte för att den blivit tillsagd att inte göra det, utan för att nyckeln inte räcker.

2. Agent-API:t kräver AGENT_API_KEY (src/agent/internal/api/middleware/auth.go). Innan dess kunde vem som helst som nådde tjänsten besluta om ett godkännande åt människor. Nu skickar Nuxt nyckeln server-side och webbläsaren ser den aldrig.

Lärdomar:
- Två nycklar med samma värde men olika scope är förvirrande. Agentens nyckel och bypass-nyckeln måste vara olika värden, annars vet man inte vilken regel som gäller.
- Allowlisten ska uttryckas som "vad agenten behöver", inte "vad agenten inte får göra". Det gör att en ny agent-funktion tvingar fram en medveten tilläggning.
- En gammal modell (att agentens egen tjänst beslutar med sin nyckel) bryter kedjan. Rätt väg är att människans session bestämmer i Plan, och att agenten bara läser resultatet.
- Testa med riktiga anrop: DELETE av ett mål med agent-nyckeln gav 200 när scope inte var active - det syntes först när koden rullats ut. En test i Go är inte samma sak som ett test i klustret.
