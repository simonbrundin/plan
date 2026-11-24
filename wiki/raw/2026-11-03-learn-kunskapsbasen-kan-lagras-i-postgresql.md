---
title: Learn: kunskapsbasen kan lagras i PostgreSQL
type: source
added: 2026-11-03
origin: internal:learn
immutable: true
---

---
title: "Learn: kunskapsbasen kan lagras i PostgreSQL"
type: note
tags: [learn, postgres, lagring, go]
---

Learn lagrar fortfarande lokalt i SQLite, men PostgreSQL-lagret är nu implementerat enligt samma gränssnitt (store.Repository), så API:t och CLI:t fungerar oförändrat mot båda.

Varför: AGENTS.md i learn-repot beskriver PostgreSQL (CNPG) och Atlas som produktionsdatabas, och kunskapsbasen ska kunna delas mellan miljöer. SQLite fungerade men kunde inte användas i klusret.

Implementation i /home/simon/repos/learn/src/go:
- internal/store/postgres.go - Postgres-implementation av store.Repository. Samma tabeller och semantik som SQLite-lagret. Skillnader: JSONB för nyckelord/citat/alternativ/idébasis, TIMESTAMPTZ för tidsstämplar, RETURNING id i stället för LastInsertId, ON CONFLICT DO NOTHING för relationer.
- internal/store/open.go - OpenFromEnv väljer PostgreSQL när DATABASE_URL är satt, annars SQLite.
- cmd/learn/main.go - kommandona tar store.Repository i stället för *store.Store, så båda lagren fungerar. --db styr SQLite, DATABASE_URL styr PostgreSQL.
- cmd/api/main.go - samma växling.

Lardomar:
1. Att byta från *store.Store till store.Repository i CLI:t var en mekanisk ändring (14 funktioner) men gav även ett fritt extra: Path() lades till gränssnittet så lagringen kan beskrivas i utdata.
2. Postgres-tester behöver en egen databas per test. Enklast är CREATE DATABASE via en admin-anslutning och DROP DATABASE i t.Cleanup. search_path per schema fungerar inte med database/sql-poolen eftersom sessioner växlas.
3. Testerna hoppas över när LEARN_TEST_DATABASE_URL saknas, så vanliga körningar kräver ingen databas.
4. local postgres:16-alpine på port 5432 med postgres/postgres finns på maskinen och användes för verifieringen.

Verifierat: import, list, ask och /insight via CLI och API mot PostgreSQL 16. Relationer, tidsstämplar, JSONB-kolumner och borttagning fungerar.
