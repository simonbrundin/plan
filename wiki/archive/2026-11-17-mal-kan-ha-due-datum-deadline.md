---
title: Mål kan ha due-datum (deadline)
type: note
tags: [feature, goals, go, nuxt, db, projekt:plan]
---

2026-11-17: Mål i plan kan nu ha ett due-datum (deadline).

VARFÖR: målen hade bara started/finished - ingen deadline, så det gick inte att
säga "det här ska vara klart senast X".

HUR:
- DB: ny kolumn "due" DATE (nullable) i goals. Kanoniskt schema:
  environments/base/db/schema.yaml (ConfigMap med inbäddad SQL). Atlas Operator
  applicerar den mot Neon i dev.
- Go: egen typ models.DueDate (time.Time med Marshal/Unmarshal som "YYYY-MM-DD")
  + NullableDueDate i UpdateGoalRequest, så att PATCH kan sätta (värde),
  rensa (null) eller lämna orört (fältet saknas). Ogiltigt datum ->
  repository.ErrInvalidDueDate -> HTTP 400.
- API: GET /goals och GET /goals/:id returnerar "due" (alltid, även null).
  PATCH /goals/:id med {"due": "2026-11-20"}.
- Frontend: tangent 'c' + kalenderknapp i målsidan (/goal/:id) och i listan
  (/goals). Datumväljaren är ett native <input type="date">; datumet sparas
  direkt vid val, X-knappen rensar. Datumet visas som badge (röd = förgått,
  gul = idag) på målsidan, i grundmålslistan, bland undermål och på
  prioriteringssidan. Ny util: src/nuxt/app/utils/dueDate.ts.

VÄRKTIGT: due är ett KALENDERDATUM, inte en tidsstämpel - därför egen typ och
DATE-kolumn istället för TIMESTAMPTZ. new Date("2026-11-20") tolkas som
UTC-midnatt och kan visa fel dag i tidszoner bakom UTC; utilen bygger därför
datumet lokalt.

FALLER: environments/pr/template/db/schema.sql är en gammal, oanvänd kopia av
schemat (saknar t.ex. eternal/order och kustomize-sökvägen
../../../../common/db finns inte) - den uppdaterades inte.

VERIFIERAT: enhetstester (Go + bun), integrationstest mot Neon
(goal_repository_integration_test.go), samt manuellt i browsern: sätta datum,
byta datum och rensa via UI:t.
