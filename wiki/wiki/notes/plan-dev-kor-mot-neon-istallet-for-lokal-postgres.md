---
title: Plan dev kör mot Neon istället för lokal Postgres
type: note
tags: [neon, dev, database, plan, projekt:plan]
created: 2026-11-16
updated: 2026-11-16
status: active
---

Projekt: plan

Dev-miljön (kind-kluster, namespace plan) kör inte längre mot CNPG-klustret plan-db, utan mot Neon (Lakebase Postgres), projekt fragrant-pine-27757562, branch dev (br-misty-frost-aldds5s5), databas neondb.

Varför: dev-klustret ska slippa bära en Postgres, och samma branch-utvecklingsflöde som används i stk-repot (där dev redan kör mot Neon). Prod kör fortfarande mot CNPG.

Teknik: environments/dev/db/neon.env (gitignoras) -> kustomize secretGenerator -> Secret plan-db-neon med nycklarna DATABASE_URL (poolad, channel_binding=require for pgx), DATABASE_URL_UNPOOLED (direkt) och uri (direkt, UTAN channel_binding - Atlas använder lib/pq som inte stöder parametern). CNPG-klustret i base/db raderas i dev via en : delete-patch, och AtlasSchema pekas om mot plan-db-neon/uri. pgAdmin-servernsdefinitionen genereras av environments/dev/db/sync-neon-env.sh ur neon.env.

Viktigt: en deklarativ Atlas-apply raderar tabeller som inte finns i schema.yaml. Testtabeller i Neon måste tas bort manuellt.

Verifierat: kustomize build environments/prod är identisk med före (prod-overlayn orörd), dev-renderingen innehåller inga CNPG-resurser eller plan-db-referenser, AtlasSchema Applied mot Neon, go-api startar och dev-login skriver session i Neon.
