---
title: Relationer kan nu läggas till från översikten - dialogen har par-läge
type: note
tags: [pappa, relationer, ui, dialog, refaktor, projekt:pappa]
created: 2026-11-17
updated: 2026-11-17
status: active
related: [notes/relationsoversikt-i-pappa-var-ser-jag-mina-egna-relationer]
---

2026-11-10: 'jag vill ju kunna lägga till fler relationer jag har i den listan' - alltså: från relationsöversikten, utan att gå via ett kandidatkort. Tidigare var dialogen personcentrerad (prop person: Person), så det fanns ingen väg att skapa ett par från en lista.

LÖSNING: RelationDialog fick ett PAR-LÄGE. Samma komponent, två lägen:
- person-läge: 'Vem känner X?' + befintliga relationer + välj den andra
- par-läge: 'Lägg till en relation' + två personväljare
Root-elementet kan inte ha v-if="person" - då renderas ingenting i par-läge (hade jag gjort först, fångades i browsertest).

FALLGROP (riktigt bra): adminsessionsstatusen hämtades bara i person-watchen. I par-läge triggas ingen person-watch -> adminAuth false -> spar-knappen låst trots inloggning. Fix: onMounted -> refreshAdmin(). Fångades bara för att testet klickade på en REDIGERA-knapp först; i par-läge såg man bara en knapp som inte gjorde något.

FALLGROP 2: styrka/typ/anteckning låg duplicerat i två lägen -> utdrag till RelationFields.vue med defineModel. Det är samma princip som resten av repot: en källa för sanning, inte två kopior.

DUBBLETT-SKYDD: pairExists() kollar mot /api/people:s relationer (symmetriska, så ena sidan räcker). Finns relationen visas det och spar-knappen låses - i stället för att skapa en tredje rad. Par-läge är alltså inte bara en bekvämlighet utan också ett ställe där dubbletter förhindras.

Verifierat: hela flödet i browser (öppna dialogen från översikten, välja två personer, styrka 7, spara, relationen syns under 'Mina' med 'ägs av Michael Brundin'), plus att person-läget fortfarande fungerar (styrkeändring 7->9 sparas). Testrelationen raderades efteråt - databasen har 39 relationer och noll ägare.

## Relaterade

- [[notes/relationsoversikt-i-pappa-var-ser-jag-mina-egna-relationer]]

