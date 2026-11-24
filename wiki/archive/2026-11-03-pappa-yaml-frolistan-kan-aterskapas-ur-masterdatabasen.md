---
title: pappa: YAML-frölistan kan återskapas ur masterdatabasen
type: note
tags: [pappa, databas, yaml, recovery, verktyg, projekt:simon]
---

2026-11-08: data/sundsvall-it-people.yaml skadades under en textskriptning (skriptet skrev över filen utan att ta med filhuvudet). Personen återställdes ur masterdatabasen, och verktyget som gjorde det ligger nu i repot.

VARFÖR det gick att återställa: masterdatabasen (Neon/Postgres) innehåller allt som frölistan lagrar per person - external_id, titel, organisation, källor, kompetenser, kontakt, noter, erfarenhetsgrund. YAML-filen är frölista, DB är sanningen som appen läser.

NYTT VERKTYG: db/regenerate_people_yaml.py skriver om people:-sektionen ur databasen och rör inte övriga sektioner (meta, legal, employers_to_work, sources, competence_taxonomy, next_steps, notes - de är redaktionella och finns bara i YAML). Verktyget är idempotent och vägrar skriva om filen om den genererade versionen inte går att tolka.

FÅLLE: I YAML flow-sammanhang ({ url: ... }) är komma, hakparenteser och klammer indikatorer. En URL med komma i sig MÅSTE citeras, annars bryts raden sönder vid parsning och källan tappar sin sista del. q()-funktionen i verktyget citeraar därför alltid värden med , [ ] { }.

LÄRDOM: (1) Redigera YAML med radbaserade operationer, aldrig med indexbaserad slicing av en hel fil - en felaktig index matchar fel sektion (t.ex. 'sources:' finns både i personposterna och som egen sektion). (2) Skriv om YAML via en parser, inte via strängsammansättning, när värdet kan innehålla separatorer. (3) Ha en återställningsväg från denormaliserad data - utan den hade 88 av 140 personer varit tappade.
