---
title: Poolad eller direkt anslutning mot Neon
type: concept
tags: [neon, postgres, pgbouncer, drift, projekt:tmp]
created: 2026-11-02
updated: 2026-11-02
status: active
related: [notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen]
---

Neon ger TVÅ endpoints för samma databas. Skillnaden är inte kosmetisk - den
avgör vilka verktyg som fungerar alls.

  poolad   ep-xxx-POOLER.aws.neon.tech   -> PgBouncer i transaktionsläge
  direkt   ep-xxx.aws.neon.tech          -> riktig Postgres-session

VAD SOM INTE FUNGERAR VIA DEN POOLADE ENDPOINTEN:
- SET / RESET (t.ex. search_path) - sessionstillstånd försvinner när
  transaktionen är klar.
- SQL-nivå PREPARE / EXECUTE / DEALLOCATE.
- pg_dump / pg_restore (dumpen är sessionsbunden).
- LISTEN/NOTIFY, session-nivå advisory locks, WITH HOLD cursor.
- Tillfälliga tabeller som ska överleva transaktionen.

DET FUNGERAR: vanliga frågor, DML, DDL, protokollnivå-prepared statements
(PgBouncer 1.22+ stöder dessa), upp till 10 000 samtidiga klienter.

REGEL: applikationens frågetrafik går via den poolade URL:en. Allt som är
sessionsbundet - migrationer, dump, restore, replikering - går via den
direkta URL:en.

FÄLLOR:
1. Werktyg som bara läser EN connection string (t.ex. Prisma Migrate, Django
   migrate, Rails db:migrate) körs fel om de pekar på poolad URL. Ge dem
   DATABASE_URL_UNPOOLED.
2. psycopg 3 avslutar prepared statements med SQL-nivå DEALLOCATE, vilket
   PgBouncer inte stödjer. Mot poolad URL: sätt prepare_threshold=None.
3. Efter en schemaändring (DROP/CREATE) kan en klient träffa en gammal
   plan-cache i PgBouncers backend och få "cache lookup failed for type N".
   Det är transient - koppla om och försök igen.
4. Scale-to-zero (5 min) ger kallstart: första frågan kan möta "the database
   system is starting up". Hantera med retry.
5. Filen .env.local från Neon CLI innehåller lösenordet i klartext - den
   får aldrig committas.

SE OCKSÅ: notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen

## Relaterade

- [[notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen]]

