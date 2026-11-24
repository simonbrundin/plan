---
title: pappas dev-miljö kör mot Neon i stället för in-kluster-Postgres
type: note
tags: [neon, kubernetes, tilt, deployment, tls, projekt:pappa]
---

2026-11-02: environments/ i pappa var orörd boilerplate från repo-template (go-api,
CNPG-Postgres, Atlas-operator, pgadmin, users-schema) - den pekade på en Go-tjänst
och ett schema som inte existerar i pappa. Dev byggdes om till att använda Neon.

VAD SOM ÄNDRADES:
- environments/dev/neon/ (ny): secretGenerator ur neon.env -> Secret
  pappa-neon-db (DATABASE_URL poolad + DATABASE_URL_UNPOOLED). neon.env genereras
  av environments/neon-env.sh ur src/neon/.env.local - gitignoras, aldrig i git.
- environments/base/nuxt-deployment.yaml: NUXT_DATABASE_URL ur secreten
  (runtimeConfig.databaseUrl i nuxt.config.ts). NUXT_PUBLIC_API_BASE bort.
- Borttaget: dev/db (CNPG Cluster + app-secret), dev/pgadmin (+ vendored chart),
  base/go-api-*.yaml, base/db (Atlas + repo-template-schemat), prod/db, pr/template.
  Prod och PR har ingen databas i klustret - databasen är Neon överallt.
- environments/pr/ (ny neon/ + fixad httproute): backendRef pekade på servicen
  "app" som inte finns - nu "nuxt-app-service". Prod har samma fel (fixat).
- environments/prod/neon/README.md: hur Secret pappa-neon-db skapas externt
  (kubectl/Flux) - aldrig ur fil i repot.
- environments/dev/Tiltfile: ingen go-api, ingen pgadmin, synkar neon.env.
- src/nuxt/Dockerfile (ny): node:22-alpine + Nitro-server. pappas Nuxt har
  server-side rendering och /api/people, så repo-templates nginx-SPA fungerar
  inte - kräver Node.

TLS (tillägg 2026-11-03):
- src/nuxt/server/utils/db.ts: ssl 'require' -> 'verify-full'. I postgres.js
  betyder 'require' kryptering UTAN certifikatkontroll; 'verify-full' kollar
  också värdnamn. resolveSsl() slår på plaintext bara för localhost.
- src/neon/pappa_db.py: libpq-sslmode 'require' -> 'verify-full' + sslrootcert
  mot en CA-fil (/etc/ssl/certs/ca-certificates.crt m.fl. eller PAPPA_SSLROOTCERT).
  libpq:s "system" fungerar inte på alla plattformar - peka ut filen explicit.

VARFÖR:
1. Dev ska använda samma databas som resten av pappa. Med en in-kluster-Postgres
   hade man behövt dubbel datamodell och dubbel drift.
2. Secreten genereras ur src/neon/.env.local, den enda källan för
   anslutningssträngen. Ingen andra handhållen kopia av lösenordet.
3. go-api fanns inte (ingen src/go i pappa) och Atlas-schemat var
   repo-templates users-tabell - inget av det hade någon funktion.
4. Certifikaten ska kontrolleras: databasen innehåller personuppgifter, och
   "krypterat men okänt motpart" räcker inte för GDPR-syftet.

VERIFIERAT: kustomize build base/dev/prod/pr går igenom; docker build av
src/nuxt fungerar; container med enbart NUXT_DATABASE_URL svarar
/api/people med 88 personer; npm run db:import (idempotent) och
nuxi typecheck går igenom mot Neon med verify-full.

KVAR: environments/pr/kustomization.yaml skrivs av GitHub Actions - om den
filen regenereras måste resursen ./neon finnas med, annars saknas secreten.
