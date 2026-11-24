---
title: Mål kan ha startdatum - syns i /priority först när datumet passerat
type: note
tags: [feature, goals, priority, go, nuxt, db, projekt:plan]
---

2026-11-20: Mål kan nu ha ett **startdatum** (goals.start_date, "YYYY-MM-DD"):
det datum då målet tidigast får dyka upp i /priority.

VARFÖR: det fanns ingen "tidigast synlig"-mekanism. started säger att arbetet
började, due säger när målet ska vara klart, recurrence_anchor säger när en
serie startade - men inget sade "det här målet får inte prioriteras ännu".
Ett mål som ska aktiveras i framtiden syntes i listan direkt (bara filtrerad
av "påbörjad"), vilket gjorde listan ful och gav sken av att jobbet borde
göras nu.

HUR:
- DB: ny nullable kolumn start_date DATE på goals (kanoniskt schema:
  environments/base/db/schema.yaml).
- API: PATCH /api/goals/:id med {"start_date": "2027-01-01"} (null rensar).
  Filtreringen sker i SQL (GetPrioritizedGoals): start_date IS NULL OR
  start_date <= CURRENT_DATE - alltså i servern, inte i klienten, så regeln
  gäller för listan, agenten och andra API-anrop.
- UI: tangenten B på målsidan, i /goals och i /priority. Framtida datum visas
  i blått. Ångras med Z.

VIKTIGT: requestfältet start_date betyder NU målets startdatum. Seriens
startdatum (recurrence anchor) har bytt namn till recurrence_start_date i
create/update - annars hade samma fält haft två betydelser. API-klienter
som sätter upprepningens startdatum måste uppdatera fältnamnet.

FALLER: i headless Chromium (Playwright) commit-tryck i ett date-input kan
skicka "idag" i stället för det valda datumet - det är en browser-quirk vid
programmatiskt satta värden, inte appfel (verifierat: mänsklig tangenttryck
fungerar, change-eventet skickar rätt värde).

SPEC: specs/goal-start-date.md
