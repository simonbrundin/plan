---
title: Agent-api crashade i Tilt: saknad DATABASE_URL mot CNPG-databasen i agents-dev
type: note
tags: [bug, tilt, dev, agent-api, database, plan-api, projekt:agents]
created: 2026-11-18
updated: 2026-11-23
status: active
related: [notes/agent-granssnittet-nuxt-saknades-helt-i-tilt-routes-forsvann-i-refaktoreringen-677b876]
---

Symtom: agent-api-podden i default-ns (namespace dev) gick i CrashLoopBackOff. Loggen: 'Failed to connect to database: failed to ping database: failed to connect to user=postgres database=software_agent: hostname resolving error: lookup postgres on 10.96.0.10:53: server misbehaving'.

Rotorsak: agentens deployment (blob-manifestet i environments/dev/Tiltfile) satte ingen DATABASE_URL. Go-koden (src/agent/internal/api/db/db.go) faller då tillbaka på postgres://postgres:postgres@postgres:5432/software_agent - en värd som inte finns i klustret. 'server misbehaving' (SERVFAIL) istället för NXDOMAIN kom från CoreDNS i kind-klustret.

Fix: environments/dev/Tiltfile läser nu URI:en ur environments/dev/db/app-secret.yaml (samma secret CNPG-klustret använder) via read_yaml och sätter den i agent-config ConfigMap:en. Datasabasen provisioneras av en local_resource 'agents-db' som kör kubectl apply (aldrig delete) på db/ - så en befintlig dev-databas inte raderas när Tilt uppdaterar resurser.

Varför: agenten har sin egen databas (conversations, messages, checkpoints, goals) i agents-dev; Plan har sin egen (Neon). Agenten pratar med Plan via HTTP, inte via DB.

Bonus-fynd: agentens Plan-klient var trasig mot Plan's nuvarande API. (1) paths: klienten anropade /api/goals och /api/statuses, men Plan serverar under /api/v1 -> 404. Fix: src/agent/internal/planclient/client.go använder nu /api/v1/goals och /api/v1/statuses. (2) PLAN_API_URL pekade på http://plan-go.plan:8080 - en tjänst som inte finns (den heter go-api-service.plan). (3) PLAN_WEB_URL pekade på http://plan-nuxt.plan:3000 (tjänsten heter plan-service, port 80).

Viktigt: Plan's API-nyckelautentisering fungerar inte just nu. api-gruppen i plan kör APIKeyAuth och sedan AuthMiddleware; AuthMiddleware kräver alltid en sessiontoken och avbryter med 401 även när API-nyckeln var giltig. I dev fungerar i stället dev-token-formatet 'user_<id>' i Authorization: Bearer-headern (Plan: AuthMiddleware). Tiltfile sätter PLAN_API_KEY=user_1. En rad 'agent-dev' med sha256-hash lades in i Plans Neon-databas (api_keys, user_id=1) som väg att använda när API-nyckel-fixen landar.

Verifierat: agent-api-podden 1/1 Running, /api/v1/health svarar 200, loggen visar 'Found 0 goals with status=define' (inga 404/401 kvar).

## Relaterade

- [[notes/agent-granssnittet-nuxt-saknades-helt-i-tilt-routes-forsvann-i-refaktoreringen-677b876]]

