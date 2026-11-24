---
title: Enter skapar alltid nytt mål - piltangenterna markerar befintligt mål i sökfält
type: note
tags: [bug, ui, goals, tangentbord, projekt:plan]
---

PROBLEM: När man söker efter ett mål (undermål/förälder på målsidan, eller föräldersökningen i snabbfältet) och trycker Enter valdes det första sökresultatet - ett befintligt mål. Det gick därför inte att skapa ett nytt mål med samma namn som ett som redan finns (t.ex. "Äta" under "Hälsa"). Användaren kunde inte välja nytt läge utan att söka på ett annat sätt.

LÖSNING: Sökfälten har nu en markering (index -1 = inget markerat).
- Enter skapar ALLTID ett nytt mål med söktexten som namn så länge inget resultat är markerat.
- ↑/↓ markerar ett befintligt mål i listan; Enter väljer då det markerade målet (t.ex. som undermål/förälder).
- Markeringen nollställs när söktexten ändras.

Beroenden: src/nuxt/app/utils/searchHighlight.ts (ren logik, enhetstestad med bun test), används av både pages/goal/[id].vue (förälder- och undermålssökning) och components/QuickAdd.vue (föräldersökningen i snabbfältet). I snabbfältet skapar Enter i huvudfältet nu direkt målet om ingen föräldrasökning är påbörjad.

Verifierat headless (chromium + CDP, lokal Postgres + Go-API på port 8091): Enter skapade nya mål med samma namn, ↑+Enter använde det befintliga målet, och snabbfältet skapade mål med markerad förälder.

VARNING (test): "Middag" tolkas av den naturliga språkparsern som en TID - snabbfältet skapar då en händelse, inte ett mål. Använd andra ord vid test.
