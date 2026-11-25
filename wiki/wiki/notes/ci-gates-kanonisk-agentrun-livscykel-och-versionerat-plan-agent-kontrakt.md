---
title: CI-gates, kanonisk AgentRun-livscykel och versionerat Plan-agent-kontrakt
type: note
tags: [agent, ci, kontrakt, databas, projekt:plan]
created: 2026-11-25
updated: 2026-11-25
status: active
related: [notes/agent-scope-i-prod-allowlistan-maste-matcha-kontraktet]
---

2026-11-25: mål 497 (fas 1) färdigställdes. Delmålen 499 (checkpoint-orchestratorn startar),
500 (hela testsviten i CI) och 501 (AgentRun-statusar) är markerade klara, fasmålet 498
därmed också, och 522 (gemensamma kontrakt) har en implementation. Kvar för 497 är bara
proddriften.

499 - orchestratorn startar och tillståndet syns. StateStoren är CheckpointStore (Postgres)
när databasen finns, så checkpointplanen överlever en omstart i dev och prod; minneslagring
är fallback. Initieringsfelet exponeras i /api/v1/health och /api/v1/status
(orchestrator.initialized, orchestrator.error). Varför: utan orchestratorn registreras
inga workflow-endpoints, och det ska gå att se att de saknas utan att läsa loggar.

500 - testerna kördes inte alls i CI. Nuxt-enhetstesterna (233 st) är skrivna med "bun:test"
och kördes bara lokalt - CI körde cucumber-flödena. De körs nu via "bun run test:unit".
Typecheck-gaten är en BASLINJE-gate: gränssnittet har 170 ärvda typfel i app/ och server/,
så gaten jämför antalet fel per fil mot src/nuxt/typecheck-baseline.json och faller bara på
nya fel (node scripts/typecheck.mjs --update skriver om baslinjen). Plan-API:t fick en
golangci-lint-gate, och de 12 linterfynden rättades så att gaten passerar från start.
Skälet till baslinjegaten: att kräva att allt är rent vore ett eget projekt, och en gate
som alltid är röd är inte en gate.

501 - statusarna heter nu waiting_approval och succeeded (tidigare blocked och done).
Målet 501 angav uttryckligen de namnen, så de byttes överallt i samma commit: API, agent,
gränssnitt, databas och docs. Lärdom: databasen har en CHECK-constraint
(agent_runs_valid_status) som låser värdena - en runtime-migrering som bara gör UPDATE
failar. Constrainten måste bort först, sedan byts raderna, sedan sätts constrainten
tillbaka, allt i en implicit transaktion. Constrainten ligger också i Atlas-schemat
(environments/base/db/schema.yaml och environments/pr/template/db/schema.sql) - båda
måste uppdateras, annan konflikt. Faser, tillstånd och övergångar är nu kanoniska i
src/go/internal/models/agent_run.go och exponeras för klienter via
GET /api/v1/agent-run-lifecycle.

522 - kontraktet mellan Plan och agenten ligger i src/contracts/plan-agent.yaml
(version 1) och är versionsstyrt. Båda sidor testar mot samma fil: Plan
(src/go/internal/contracts) jämför modellen och OpenAPI-specifikationen, agenten
(src/agent/internal/planclient) jämför sina statuskonstanter och sökvägar. Sökvägarna
ligger nu i paths.go som namngivna konstanter - annars går det inte att jämföra dem.
Testet hittade att Swagger-annotationerna för mål, kalendrar, statusar och api-nycklar
låg på /api/... medan routen är /api/v1/... - hela specifikationen är regenererad.

Prod: agenten är inte längre borttagen i prod-overlayson (hemligheterna är krypterade i
environments/prod/secrets/agent-*.enc.yaml och de överflödiga patcharna är borttagna).
Kustomize-rendering av environments/prod är verifierad: 25 resurser, bland dem agentens
deployment, CNPG-klustret, RBAC och de fyra agenthemligheterna. Kvar i prod: köra
environments/prod/secrets/plan-api-key.sql (hashen av agentens API-nyckel) och låta Flux
synka - vilket kräver att klustret är igång.

## Relaterade

- [[notes/agent-scope-i-prod-allowlistan-maste-matcha-kontraktet]]

