---
title: Agentmål: flaggan blev levande (MVP)
type: note
tags: [feature, agent, bug, session, projekt:plan]
created: 2026-11-24
updated: 2026-11-24
status: active
---

Plan hade redan goals.agent men ingen agent läste flaggan. MVP: agent_runs-tabell i Plan, GET /api/v1/goals?agent=true&status=, GET/POST/PATCH /api/v1/agent-runs, Nuxt-vy /agent, och agent-worker i agents-repot som specificerar agentmål och flyttar dem define->plan.

Fyra buggar hittades på vägen: (1) API-nyckelautentiseringen var trasig - AuthMiddleware kör även när APIKeyAuth redan satt userID, så alla agent-anrop avbröts med 'Authorization header required'. (2) PATCH /goals/:id ignorerar status_id - status ändras via PATCH /goals/:id/status; agents' planclient använde fel endpoint. (3) status_updates.changed_at (timestamptz) skannades in i en string - varje statusändring gav 500. (4) provider.CreateFromEnv satte endpoint till tom sträng när env-variabeln saknades - MiniMax-anropet gick inte igenom.

Lärdom: när agenten ska agera på data i Plan måste flaggan agent=true vara enda ingången, och agent_run ska ligga i Plan (inte i agentens egen DB) så UI:t fungerar även när agent-tjänsten är nere.

Miljö: lokal egress till Neon (psql och Go/pgx) börjar timeouta - kör DB-anrop via kubectl exec in i klustret (kind-dev, ns plan) mot hemligheten plan-db-neon (nyckeln 'uri' använder lib/pq och saknar channel_binding).
