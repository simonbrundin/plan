---
title: Erfarenhetsfilter i pappa: it_experience_years med spårbart grund
type: note
tags: [databas, gdpr, sourcing, filter, projekt:pappa]
created: 2026-11-03
updated: 2026-11-03
status: active
related: [notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen, notes/nuxt-4-webbapp-for-pappa-laser-personerna-ur-neon, notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken]
---

2026-11-03: önskemål om ett filter som sorterar bort personer som inte jobbat
med IT i minst tre år. Problemet: det finns ingen uppgift om erfarenhetslängd
i modellen - person_employment har tomma valid_from/valid_to, source.accessed_at
är importdatum (värdelöst), och bara 23 av 88 har notes.

LÖSNING: fältet it_experience_years (int, nullable) + it_experience_basis
(text) på person. Värden härleds i importen i tre steg:
1. explicit  - it_experience_years i YAML (källbelagt) - 10 personer
2. notes     - regex ur notes ("34 års erfarenhet") - 0 just nu
3. seniority - undre gräns från rollnivå (cxo/chef 10, lead 6, senior 5,
               mid 3, junior 0) - 75 personer
Okänt (null) - 3 personer.

VARFÖR SÅ HÄR:
1. Grunden sparas i it_experience_basis. Ett filtervärde utan källa kan inte
   granskas - och i ett GDPR-projekt måste varje personuppgift ha spårbar
   grund. Rollnivå-värdet är en gissning (undre gräns), inte ett facit, och
   det märks som "(uppskattat)" i UI:t.
2. Okänt räknas ALDRIG som "3+ år". Filtret får inte anta - en person med
   okänt läge kan ha 20 år. UI:t redovisar hur många som döljs på grund av
   just det.
3. Heltalsfält (år) i stället för boolean: "minst 3 år" ska gå att justera
   till 5 eller 10 utan ny migration.

EFFEKT (88 personer): 3+ år -> 84, 5+ år -> 78, 10+ år -> 20.

TEKNISKT: kolumnerna lades till med alter table ... add column if not exists
eftersom 'create table if not exists' inte ändrar befintliga tabeller - och
det MÅSTE stå före indexet på it_experience_years. v_people_full fick de nya
kolumnerna sist i listan ('create or replace view' kan bara appenda).

Relaterat: [[notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen]]
[[notes/nuxt-4-webbapp-for-pappa-laser-personerna-ur-neon]]

## Relaterade

- [[notes/pappa-anstallningshistorik-i-cv-format-fix-av-kompetensordboken]]
- [[notes/masterdatabas-i-postgres-schema-import-och-drift-for-it-rekryteringen]]
