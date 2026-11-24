---
title: Relationer mellan kandidater i pappa - symmetriska par med styrka 1-10
type: note
tags: [pappa, feature, relationer, rekrytering, databas, gdpr, projekt:pappa]
created: 2026-11-17
updated: 2026-11-17
status: active
related: [notes/detaljvy-for-kandidat-i-pappa-gemensamma-vagar-mellan-kandidater, notes/pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning, notes/rekryteringspipeline-board-per-person-sourcing-som-attribut]
---

2026-11-10: Kandidater kan nu ha relationer till varandra ('känner varandra') med styrka 1-10.

BESLUT:
- Symmetrisk relation, ett par per rad i tabellen person_relation. Raden lagras alltid med lägsta uuid:en först, och CHECK-villkoret (person_id < related_person_id) gör att en bakvänt rad inte går att skriva. Utan canonicaliseringen skulle 'A känner B' och 'B känner A' kunna ligga som två rader.
- Styrka 1-10 (heltal) - 1 = känner igen, 10 = nära vän/familj.
- Skrivning kräver adminsession (samma modell som pipen, kravprofiler, parkering) och loggas i audit_log.
- Relationen raderas när någon i paret raderas (cascade delete) - den är personuppgifter om tredje part.

VARFÖR INTE AUTOMATISK HÄRLEDNING: inga relationer härleds ur 'samma bolag' eller 'samma konferens'. Det vore en hypotes, och en hypotes som visas som fakta är värre än inget - samma princip som i dubblettslagningen (automatiska felmatchningar är inte acceptabla, det är personuppgifter). De 39 relationer som finns i frölistan kommer ur det OFFENTLIGA BOLAGSREGISTRET: personer som sitter i samma styrelse känner varandra. Det är dokumenterat, inte gissat.

YAML: toppnivå 'relations:' med a/b som yaml-id, strength, type, note. Importen lägger bara in nya par (on conflict do nothing) så att en relation som justerats i appen inte skrivs över, och tar inte bort par som försvunnit ur YAML - de gallras med personerna.

UI: chip '💞 n kända' på personkortet + dialogen (RelationDialog.vue). Personer från samma organisation listas först i väljaren, men beslutet är människans.

FALLGROP: uuid:er jämförs som bytes i Postgres = samma sak som lexikografisk jämförelse av kanoniska hexsträngar, så klientens normalizeRelationPair och DB:s CHECK landar i samma rad. Det är inte självklart - därför finns det egna enhetstester (src/nuxt/server/tests/relations.test.ts).

## Relaterade

- [[notes/detaljvy-for-kandidat-i-pappa-gemensamma-vagar-mellan-kandidater]]
- [[notes/rekryteringspipeline-board-per-person-sourcing-som-attribut]]
- [[notes/pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning]]
