---
title: Återkommande mål: samma målobjekt med upprepningsregel
type: note
tags: [feature, goals, recurrence, go, nuxt, db, projekt:plan]
created: 2026-11-18
updated: 2026-11-18
status: active
---

Återkommande mål (t.ex. 'Gymma varannan dag') är INTE en ny entitet - de är vanliga mål i goals-tabellen med en extra RRULE-sträng. Förekomsterna loggas i en ny tabell goal_completions.

VARFÖR: en separat routines-tabell + /rutiner-vy hade duplicerat hela målmodellen (titel, ikon, platser, prioritet, relationer) och brutit mot kravet att återkommande mål ska synas på samma ställen som vanliga mål.

NYA KOLUMNER I goals: recurrence_rule (RRULE), recurrence_mode (schedule|push), recurrence_anchor (seriens startdatum), recurrence_next (nästa oslutförda förekomst). Ny tabell goal_completions (goal_id, occurrence_date, completed_at) med unik constraint på (goal_id, occurrence_date).

VINST: 'finished' sätts bara när serien tar slut - så ett återkommande mål ligger kvar i listorna men är utgråat (recurrence_next > idag) tills nästa förekomst blir aktuell. Vanliga mål fortsätter försvinna när de slutförs. Listfiltreringen (isUnfinishedGoal) behövde därför INTE ändras.

VARNING (design som lärde sig): om 'recurrence_next' pekar på nästa förekomst och man checkar av, så flyttas pekaren - att bara anropa 'slutför' igen slutför NÄSTA förekomst, inte samma. Idempotens kräver därför att klienten skickar occurrence_date (vilken förekomst den avser), och servern svarar med en no-op om den redan är slutförd. Unik constraint + ON CONFLICT DO NOTHING är inte nog - servern måste också kontrollera att den anropade förekomsten är den aktuella.

SCHEMALÄGGNING: schedule håller serien fast relativt recurrence_anchor (förfallna hoppas över), push räknar nästa förekomst från slutförandedagen. Väljs per mål, inte globalt.

NATURLIGT SPRÅK: quick-add i mål-läge plockar ut frasen ur titeln ('Gymma varannan dag' -> titel 'Gymma', regel FREQ=DAILY;INTERVAL=2). I automatiskt läge blir en upprepning fortfarande en kalenderhändelse.

TEKNISK HÅLLNING: rrule-go (v1.8.2) används även för mål, med explicit DTSTART - annars använder biblioteket time.Now().UTC() och förekomstberäkningen blir tidsberoende. rrule-go accepterar INTERVAL=0, så det valideras separat i src/api/internal/recurrence.

FALLER: dev-databasen (Neon) svarar på IPv6 men Go-pgx timeout:ar - använd options=endpoint%3D<endpoint-id> i anslutningssträngen för att komma förbi SNI-kravet. Enhetstester (bun test) kan inte importera '~/utils/...' (aliasen finns inte i bun) - använd relativa importer med .ts-extension i delade utils.
