---
title: Kalenderns detaljnivåer: fem nivåer och lane-packning
type: note
tags: [kalender, feature, zoom, ui, projekt:plan]
---

Den vertikala kalendern har nu fem detaljnivåer som väljs utifrån hur mycket
tid som syns: full (kort med klockslag, <48h), compact (kort utan klockslag,
48h-7d), daily (prick per dag, 7d-45d), monthly (prick per månad, 45d-3 år)
och yearly (prick per år). Klick på en prick hoppar in en nivå.

**Varför gränserna ser ut så de är:** de är satta efter hur många markörer
som får plats i vyhöjden, inte efter kalendiska gränser. En prick per dag i
ett årfönster skulle kräva 365 prickar som varken går att läsa eller klicka
på. Därför är scaleLevel (kalendisk tidsskala) och detailLevel (hur mycket en
händelse visas) två oberoende koncept.

**Överlapp löstes med lane-packning:** händelser och prickar som ligger nära
varandra i tid flyttas åt sidan i kolumner, som en kalendervy gör. Kolumnerna
 delas bara mellan de som faktiskt överlappar (assignClusteredLanes), så en
ensam händelse får hela bredden. Avståndet mäts i pixlar via
ResizeObserver på viewporten, annars blir resultatet olika på mobil och
desktop.

**Subtilitet som kostade tid:** intervallen i packningen ska utökas med
kortets höjd, inte med avståndet mellan korten. Annars tror algoritmen att
kort en timma ifrån varandra överlappar och de delas onödigt bredd.

Verifierat i riktig browser (desktop + mobil): inga överlapp på någon nivå,
hela drill-down-kedjan fungerar, inga konsolfel. Logik i
src/nuxt/app/utils/calendarDetail.ts med Cucumber-tester i
features/calendar-detail.feature.
