---
title: Använda agenter i Plan
type: concept
tags: [agent, gränssnitt, human-gate, nuxt, projekt:plan]
created: 2026-11-26
updated: 2026-11-27
status: active
related: [notes/dokumentationssidor-i-granssnittet-docs, notes/human-gate-som-teknisk-grans-scope-pa-api-nyckeln, notes/kortkommando-s-byta-malfas-status-och-bulkmarkera-klart]
---

# Använda agenter i Plan

Så här använder en människa agenterna i Plan-gränssnittet (Nuxt, `src/nuxt`).

## Steg för steg

1. **Markera ett mål som agentmål.**
   - Målsidan (`/goal/:id`) och mållistan (`/goals`): **mellanslag + a**
     (leader-sekvens) på det markerade målet. Enbart `a` går inte - den
     tangenten används redan för att flytta markören vid titelredigering.
   - `z` ångrar agent-växlingen (`undoAgent`).
   - Klicka på agent-ikonen (`AgentBadge`) i mållistan.
   - Alla tre vägar togglar samma flagga: `goals.agent`.

2. **Öppna målsidan.** När flaggan är satt finns en **agent-dock** längst till
   höger i fönstret (knappen med gnomen i målraden, eller **mellanslag + v**).
   Docken har tre flikar, så att bara en del syns åt gången:
   - **Chatt** - "Starta chatt", skriv meddelandet (Enter skickar), agenten
     svarar i strömmande form och föreslår en definition. Förslaget visas som
     ett godkännande med **Godkänn** / **Avslå** (+ återkoppling).
   - **Körning** - faser (`Specificerar`, `Planerar`, `Utför`, `Verifierar`,
     `Klar`), tillstånd, händelser och **Avbryt** för en pågående körning.
     Historiken är en tidslinje med de senaste händelserna först.
   - **Plan** - dela upp målet i verifierbara checkpoints och
     godkänn/avslå planen. Varje checkpoint passerar en gate (beteende, UI,
     adversarial, human).

   Docken ligger fast i fönstret: på breda skärmar täcker den kalenderspalten,
   på små skärmar lägger den sig över sidan. Målets innehåll flyttas inte
   undan, så undermålen är alltid nåbara utan att scrolla.

3. **Översikt.** Sidan `/agent` listar alla agentmål med den senaste körningens
   fas, tillstånd och sammanfattning. Sidan uppdateras var femte sekund när den
   är synlig. Länken dit finns på startsidan ("Agentmål") och via `/agent-cost`
   för kostnad och budgetgränser.

## Vad som händer under huven

`goals.agent = true` -> agentens scheduler plockar upp målet -> `AgentRun` i
Plan med status `queued` -> jobbcontroller skapar ett Kubernetes Job
(`agent-worker-run-<id>`) -> workern arbetar i en isolerad klon ->
agent/goal-<mål>-run-<körning>-branch -> forslaget testas av agenten självt ->
human gate -> kostnad rapporteras till `agent_run_usage`.

## Var agenttjänsten måste vara igång

- Dev: `kind-dev`, namespace `plan` - `agent-api`, `plan-agent-db`, `go-api`,
  `nuxt-app` (via `tilt up`).
- Utan agent-API:t fungerar målsidan, men chatten, checkpoint-gates och
  körningshistorik svarar inte.
- Agent-API:t rapporterar `orchestrator.initialized` och `orchestrator.error`
  i `/api/v1/status` och `/api/v1/health` - fel syns utan loggar.
- I prod deployas agenten av Flux; hemligheterna ligger krypterade i
  `environments/prod/secrets/agent-*.enc.yaml`.

## Behörighet (human gate)

Agentens API-nyckel har scope `agent`: den får arbeta med mål, körningar och
godkännanden men får inte **besluta** om ett godkännande, radera data eller
hantera API-nycklar. Beslutet fattas alltså alltid av en människa via Plan.

## Relaterat

- `docs/agent-platform.md` - hela plattformsbeskrivningen och kontraktet.
- `src/contracts/plan-agent.yaml` - kontraktet mellan Plan och agenten.

## Relaterade

- [[notes/kortkommando-s-byta-malfas-status-och-bulkmarkera-klart]]
- [[notes/dokumentationssidor-i-granssnittet-docs]]
- [[notes/agent-dock-pa-malsiden-ersatter-de-tre-agentpanelerna]]
