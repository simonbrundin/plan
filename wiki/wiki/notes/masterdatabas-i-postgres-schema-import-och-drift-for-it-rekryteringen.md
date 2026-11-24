---
title: Masterdatabas i Postgres: schema, import och drift för IT-rekryteringen
type: note
tags: [databas, postgres, sourcing, gdpr, projekt:pappa]
created: 2026-10-27
updated: 2026-11-03
status: active
related: [notes/erfarenhetsfilter-i-pappa-it-experience-years-med-sparbart-grund, notes/nuxt-4-webbapp-for-pappa-laser-personerna-ur-neon, notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken, notes/pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning, notes/pappa-kor-nu-pa-neon-neon-ts-env-local-och-poolad-direkt-wiring, notes/pappa-yaml-frolistan-kan-aterskapas-ur-masterdatabasen, notes/rekrytering-av-it-personal-i-sundsvall-marknadsbild-och-frolista]
---

Projekt 'pappa' har nu en riktig masterdatabas i stället för bara en YAML-fil.
2026-10-27: schema + import + backup byggda och testat mot en riktig Postgres 16.

VAD SOM BYGGDES (/home/simon/repos/pappa/db/):
- schema.sql: person, organization, person_employment (historik), competence,
  person_competence, source, person_source, watchlist + watchlist_entry,
  contact_event, dsr_request, import_run, audit_log. Vyer: v_people_full,
  v_watchlist_export, v_coverage. Funktioner: fn_find_duplicates (fuzzy),
  fn_purge_expired_personal_data (GDPR-gallring med logg).
- import_from_yaml.py: importerar de 88 personerna från YAML, idempotent,
  fuzzy dubblettkoll (tröskel 0,92 - under det flaggas, inte slås ihop).
- backup.sh / restore.sh: plain SQL + gzip, versionsagnostiskt.
- queries.sql: driftfrågor (dubblettkoll, staleness, export, gallring).
- docker-compose.yml: lokal devcontainer.

VARFÖR SÅ HÄR:
1. Proveniens i datamodellen, inte bara i ett fält. En person utan källa
   ska inte finnas i databasen - annars kan vi inte svara på "varifrån kommer
   det här?" när GDPR-frågan kommer.
2. Anställningshistorik i egen tabell. Positioner förflyttas; den ursprungliga
   YAML:en har redan "positioner från äldre källor kan vara inaktuella".
   source.accessed_at gör det mätbart istället för antaget.
3. GDPR inbyggt i schemat (retention_until, do_not_contact, contact_event,
   dsr_request, audit_log) - inte i ett separat papper. Radering loggas även
   efter att personen raderats, så att vi kan visa att begäran hanterats.
4. Dubblettkoll som databasfunktion, inte som manuellt minne. Vår egen
   import hittade t.ex. "Tommy Gustafson" mot "Tommy Gustafsson" (likhet 0,83).

VÄRDVAL - GRATIS (verifierat 2026-10-27):
- Neon Free: 1 GB lagring/projekt, 100 CU-tim/månad, scale-to-zero efter 5 min,
  inget kortkort, riktig Postgres. Rekommenderat. Välj EU-region.
- Supabase Free: 500 MB, pausar efter 1 vecka inaktivitet, Studio-GUI + REST.
  Måste ha RLS aktiverat på alla tabeller annars läser PostgREST ut anon-nyckeln.
- Aiven Free: 1 GB, bra DPA och EU-region.
Vår datamängd är ~15 MB - alla alternativ räcker gott. Gratisplanerna har ingen
automatisk backup, därför db/backup.sh via cron.

LÄRDOM UNDER BYGGET (två riktiga buggar som fick testas):
- "gunzip -c fil | grep -q mönster" under set -o pipefail ger FALSKT larm:
  grep avslutar vid träff, gunzip får SIGPIPE, pipelinen returnerar fel.
  Använd zgrep istället.
- pg_dump --format=custom skriver versionsheader från KLIENTENS version. En
  backup tagen med klient v18 kan inte läsas av leverantörens pg_restore v16.
  Därför plain SQL + gzip.

NÄSTA STEG: koppla till Teams-bevakningslistan. Neon har SQL-over-HTTP, vilket
gör att Power Automate kan POSTa direkt till databasen utan någon klient.

## Relaterade

- [[notes/pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning]]
- [[notes/pappa-yaml-frolistan-kan-aterskapas-ur-masterdatabasen]]
- [[notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken]]
- [[notes/erfarenhetsfilter-i-pappa-it-experience-years-med-sparbart-grund]]
- [[notes/nuxt-4-webbapp-for-pappa-laser-personerna-ur-neon]]
- [[notes/pappa-kor-nu-pa-neon-neon-ts-env-local-och-poolad-direkt-wiring]]
- [[notes/rekrytering-av-it-personal-i-sundsvall-marknadsbild-och-frolista]]
