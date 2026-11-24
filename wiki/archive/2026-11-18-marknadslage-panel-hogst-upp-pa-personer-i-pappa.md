---
title: Marknadsläge-panel högst upp på /personer i pappa
type: note
tags: [pappa, feature, ui, sourcing, statistik, scb, projekt:pappa]
---

Ny panel (MarknadPanel.vue) högst upp på /personer i pappa: uppskattad population (SCB Yrkesregister BAS 2024) vs. hur många som faktiskt finns i databasen.

Siffror just nu: population ~3 469 anställda i IT-yrken i Sundsvalls kommun, databasen 218 personer med ort Sundsvall (209 i poolen) = 6,3 % täckning.

VARFÖR detta byggdes: frölistan är frölista, inte population. Att se täckningen (% av den mätta populationen) direkt i UI:t gör det uppenbart att listan fortfarande är en frölista och inte ett register - det styr hur mycket ansträngning som krävs innan nästa import.

TEKNISKT:
- /api/marknad räknar per ort direkt i SQL (location + is_potential/potential_review_at), samma definition av "i poolen" som i /api/people.
- SCB-siffrorna ligger statiskt i src/nuxt/shared/utils/marknad.ts, med pekare till originalet data/scb-it-statistik-sundsvall.json - där kan siffrorna uppdateras när ny statistik hämtas.
- Viktigt förbehåll som visas i UI:t: SCB räknar ANSTÄLLDA, inte egenföretagande konsulter. Den verkliga populationen av IT-konsulter är därför högre än 3 469.

Se även notes/it-arbetsplatser-i-sundsvall-3-469-enligt-scb-yrkesregistret.
