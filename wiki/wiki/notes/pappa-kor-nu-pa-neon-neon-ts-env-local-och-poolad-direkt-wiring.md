---
title: pappa kör nu på Neon: neon.ts, .env.local och poolad/direkt-wiring
type: note
tags: [neon, databas, postgres, drift, gdpr, projekt:pappa]
created: 2026-11-02
updated: 2026-11-03
status: active
related: [concepts/neon-cli-var-neon-och-neon-ts-maste-ligga, concepts/poolad-eller-direkt-anslutning-mot-neon, notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres]
---

2026-11-02: pappa-databasen provisionerad på Neon (projekt 'konsulter',
org-long-union-19124146, projekt restless-forest-15981180, branch production,
region aws-eu-central-1). Schema, import, backup och restore kördes
första gången mot den riktiga Neon-databasen - allt verifierat.

VAD SOM ÄNDRADES I REPON (/home/simon/repos/pappa/):
All Neon-relaterat ligger i src/neon/ (flyttat dit 2026-11-02):
- src/neon/neon.ts + package.json + package-lock.json: Neon CLI:ns policyfil.
  Postgres är den enda deklarerade tjänsten (auth: false).
- src/neon/.env.local (gitignoras): DATABASE_URL (poolad, -pooler),
  DATABASE_URL_UNPOOLED (direkt), NEON_BRANCH.
- src/neon/.neon (gitignoras): CLI-kontext, projectId + branch.
- src/neon/env.sh + env.py: laddar .env.local/.env. Miljövariabler vinner över
  filen. Sätter PPA_DIRECT_DATABASE_URL till den direkta URL:en.
- src/neon/pappa_db.py: psycopg-anslutning + retry vid tillfälliga fel.
- src/neon/README.md: dokumentation av Neon-lagret.

Applikationens datalager ligger kvar i db/ och använder Neon-lagret:
- db/migrate.sh: applicerar schema.sql mot den DIREKTA URL:en.
- db/import_from_yaml.py: läser .env.local via src/neon/env.py, använder
  src/neon/pappa_db.connect(), importen körs via retry().
- db/backup.sh + db/restore.sh: använder den direkta URL:en.
- Root package.json scripts: build, db:migrate, db:import, db:backup,
  neon:plan, neon:deploy, neon:env.
- db/README.md och root README.md uppdaterade med den nya strukturen.

VEM SOM LÄSER .env.local: både db-skripten (via src/neon/env.sh + env.py)
och webappen i src/nuxt/ (server/utils/db.ts söker ../neon/.env.local relativt
till process.cwd()). Det finns alltså EN anslutningssträng i repot - flytta
den inte utan att uppdatera båda läsarna.

CLI-FÄLLA (viktigt): Neon CLI:et hittar .neon genom att söka UPPÅT från cwd
och letar neon.ts BREDVID .neon. Därför fungerar inte `neon deploy` från
repo-roten när .neon ligger i src/neon/ - det ger "Multiple projects found".
Använd `npm run neon:deploy` (som skickar --context-file src/neon/.neon) eller
gör `cd src/neon && neon deploy`. env pull/deploy skriver .env.local bredvid
.neon, alltså i src/neon/.

VARFÖR DELA MELLAN POOLAD OCH DIREKT URL:
Neons poolade endpoint är PgBouncer i transaktionsläge. Sessionstillstånd
(SET search_path), SQL-nivå PREPARE och pg_dump fungerar INTE via den.
Därför: migrationer, dump och restore går via DATABASE_URL_UNPOOLED, allt
annat via den poolade DATABASE_URL. Se concept-sidan 'poolad-eller-direkt-
anslutning-mot-neon'.

DRIVERVAL: psycopg 3 behålls (projektet har ingen ORM). Mot den poolade
endpointen sätts prepare_threshold=None - psycopg avslutar prepared
statements med SQL-nivå DEALLOCATE, vilket PgBouncer inte stödjer, och en
klientsession kan byta server-session mellan transaktioner.

LÄRDOM (riktig bugg under testet): efter att restore.sh kört (DROP + CREATE
av tabeller och enumper) fick nästa import "cache lookup failed for type
16692" - en gammal plan-cache i PgBouncers backend som pekade på en OID
som droppats. Felet gick inte att återskapa på begär, men klassen av fel är
transient. Därför körs importen via retry() (3 försök, exponentiell backoff).
Importen är idempotent, så en omkörning är säker.

LAGER SOM INTE ÄNDRATS: EU-region (eu-central-1) valdes från början, så
personuppgifterna ligger inom EU. Branchen 'production' är skyddad av
Neons filialfunktion (point-in-time restore), men gratisplanen har ingen
automatisk backup - db/backup.sh via cron är därför kvar.

## Relaterade

- [[notes/pappas-dev-miljo-kor-mot-neon-i-stallet-for-in-kluster-postgres]]
- [[concepts/neon-cli-var-neon-och-neon-ts-maste-ligga]]
- [[concepts/poolad-eller-direkt-anslutning-mot-neon]]
