---
title: "Sammanfattning: Learn-kunskapsbasen kan lagras i PostgreSQL"
type: summary
summary: PostgreSQL-lager för Learn implementerat enligt store.Repository, med DATABASE_URL som växlare och egna testdatabaser per test.
tags: [learn, postgres, lagring, go]
sources: 2026-11-03-learn-kunskapsbasen-kan-lagras-i-postgresql.md
created: 2026-11-03
updated: 2026-11-03
---

# Sammanfattning: Learn-kunskapsbasen kan lagras i PostgreSQL

## Källa

`raw/2026-11-03-learn-kunskapsbasen-kan-lagras-i-postgresql.md`

## Översikt

Learn lagrar fortfarande lokalt i SQLite, men PostgreSQL-lagret är nu implementerat
enligt samma gränssnitt (`store.Repository`), så API:t och CLI:t fungerar oförändrat
mot båda.

| Del                    | Innehåll                                                        |
| ---------------------- | --------------------------------------------------------------- |
| `internal/store/postgres.go` | Postgres-implementation av `store.Repository`           |
| `internal/store/open.go`     | `OpenFromEnv` väljer lager utifrån `DATABASE_URL`        |
| `cmd/learn/main.go`          | Kommandona tar `store.Repository` i stället för `*store.Store` |
| `cmd/api/main.go`            | Samma växling för API:t                                   |

## Skillnader mot SQLite-lagret

| Aspect         | SQLite                | PostgreSQL              |
| -------------- | --------------------- | ----------------------- |
| Listor         | TEXT med JSON         | JSONB                   |
| Tidsstämplar   | REAL (sekunder)       | TIMESTAMPTZ             |
| Nya id         | `LastInsertId`        | `RETURNING id`          |
| Dubbletter     | `INSERT OR IGNORE`    | `ON CONFLICT DO NOTHING` |

## Lärdomar

1. Att byta från `*store.Store` till `store.Repository` i CLI:t var en mekanisk
   ändring (14 funktioner) men gav även ett fritt extra: `Path()` lades till
   gränssnittet så lagringen kan beskrivas i utdata.
2. Postgres-tester behöver en egen databas per test. Enklast är `CREATE DATABASE`
   via en admin-anslutning och `DROP DATABASE` i `t.Cleanup`. `search_path` per
   schema fungerar inte med `database/sql`-poolen eftersom sessioner växlas.
3. Testerna hoppas över när `LEARN_TEST_DATABASE_URL` saknas, så vanliga körningar
   kräver ingen databas.
4. `postgres:16-alpine` på port 5432 med `postgres/postgres` finns på maskinen och
   användes för verifieringen.
