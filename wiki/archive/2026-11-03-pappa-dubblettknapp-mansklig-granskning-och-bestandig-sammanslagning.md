---
title: pappa: dubblettknapp - mänsklig granskning och beständig sammanslagning
type: note
tags: [pappa, databas, gdpr, verktyg, granskning, projekt:simon]
---

2026-11-08: Användaren bad om en knapp som försöker slå ihop dubbletter. Byggd som knapp + API + databasfunktion.

VARFÖR INTE AUTO-MERGE: dagens data innehåller tre Söderberg på Arkaden Konsult (Gert, Per, Robert) och två Mikael Nordlund-liknande par. En automatisk sammanslagning på namnlikhet skulle utplåna släktingar. Därför är kandidaterna förslag, och en människa väljer vilken post som ska behållas.

SIGNALER: fn_duplicate_candidates() rankar par på 'samma e-post' (1.0), 'samma telefon' (0.95), 'samma namn' (0.9) och namnlikhet (pg_trgm). Nuvarande topplista är 10 par, varav samtliga är falskt positiva - vilket i sig är ett resultat: databasen är ren.

TVÅ DESIGNVAL SOM SPELAR ROLL:
1. Skrivning kräver admin-token (NUXT_ADMIN_TOKEN, hemligheten pappa-admin i klustret). Läsning är publik - det är samma data som /api/people visar. En publik merge-knapp hade varit en knapp som raderar personer ur masterdatabasen.
2. Sammanslagningen registreras i tabellen person_merge. Utan det vore sammanslagningen OBESTÄNDIG: frölistan (YAML) innehåller fortfarande båda posterna, så nästa import skulle skapa tillbaka den raderade dubbletten. Importer hoppar nu över external_id som finns i person_merge.

SEMANTIK I fn_merge_people: kopplingar flyttas (källor, kompetenser, anställningar, bevakningslista, kontaktlogg), tomma fält fylls från den raderade posten men befintliga värden skrivs aldrig över, och högst en aktuell anställningsrad per person (den kuraterade raden med ref vinner över legacy-raden).

TEST: db/test_merge.sql kör sammanslagningen i en transaktion som rullas tillbaka och fäller med undantag om kopplingar inte flyttas, om två aktuella rader finns kvar, om person_merge saknas eller om spåret i audit_log saknas.

LÄRDOM: 'samma namn + samma bolag' är en FALSK signal i den här databasen - släktingar arbetar i samma konsultbolag. Namnlikhet ensam räcker aldrig; e-post och telefon är de starka signalerna.
