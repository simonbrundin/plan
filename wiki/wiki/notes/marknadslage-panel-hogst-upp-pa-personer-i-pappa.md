---
title: Marknadsläge-panel högst upp på /personer i pappa
type: note
tags: [pappa, feature, ui, sourcing, statistik, scb, projekt:pappa]
created: 2026-11-18
updated: 2026-11-18
status: active
related: [notes/it-arbetsplatser-i-sundsvall-3-469-enligt-scb-yrkesregistret]
---

Ny panel (MarknadPanel.vue) högst upp på /personer i pappa: uppskattad population (SCB Yrkesregister BAS 2024) vs. hur många som faktiskt finns i databasen.

Siffror just nu: population ~3 469 anställda i IT-yrken i Sundsvalls kommun, databasen 218 personer med ort Sundsvall (209 i poolen) = 6,3 % täckning.

VARFÖR detta byggdes: frölistan är frölista, inte population. Att se täckningen (% av den mätta populationen) direkt i UI:t gör det uppenbart att listan fortfarande är en frölista och inte ett register - det styr hur mycket ansträngning som krävs innan nästa import.

TEKNISKT:
- /api/marknad räknar per ort direkt i SQL (location + is_potential/potential_review_at), samma definition av "i poolen" som i /api/people.
- SCB-siffrorna ligger statiskt i src/nuxt/shared/utils/marknad.ts, med pekare till originalet data/scb-it-statistik-sundsvall.json - där kan siffrorna uppdateras när ny statistik hämtas.
- Viktigt förbehåll som visas i UI:t: SCB räknar ANSTÄLLDA, inte egenföretagande konsulter. De senare är 107 i Mellerstra Norrland (SCB YREG02EGF12BAS 2024) och ingår alltså inte i 3 469.

RÄTTELSE (2026-11-18): panelens första version av en konsult-uppskattning visade "50 % konsultnära" utan att skriva ut antagandena. Det var min egen indelning, inte en mätning, och användaren ifrågasatte den rätt. Nu visas i stället ett INTERVALL (≈ 1 400, dvs 1 362-1 462) med antagandena utskrivna i UI:t: alla IT-konsultbolag ligger i "information och kommunikation" (1 562), men där sitter också telekom (Telia ~390 anställda i Sundsvall) och media. Se noten om metoden.

Se även notes/it-arbetsplatser-i-sundsvall-3-469-enligt-scb-yrkesregistret.

## Relaterade

- [[notes/it-arbetsplatser-i-sundsvall-3-469-enligt-scb-yrkesregistret]]

