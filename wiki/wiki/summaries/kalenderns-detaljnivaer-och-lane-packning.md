---
title: Kalenderns detaljnivåer och lane-packning
type: summary
summary: Kalenderns fem detaljnivåer väljs efter tidsfönstrets storlek, inte efter händelsens vikt - och överlapp löses med pixelmätt lane-packning.
tags: [kalender, feature, zoom, ui, projekt:plan]
sources: [raw/2026-11-17-kalenderns-detaljnivaer-och-lane-packning.md]
created: 2026-11-17
updated: 2026-11-17
status: active
---

## Källa

`raw/2026-11-17-kalenderns-detaljnivaer-och-lane-packning.md` (origin: https://github.com/simonbrundin/plan)

## Huvudpunkter

1. **Fem detaljnivåer, valda efter tidsfönstrets storlek** - inte efter hur viktig
   händelsen är: `full` (< 48 h, kort med ikon/titel/klockslag), `compact`
   (48 h - 7 d, kort utan klockslag), `daily` (7 d - 45 d, prick per dag med
   antal), `monthly` (45 d - 3 år, prick per månad), `yearly` (> 3 år, prick per
   år). Klick på en prick hoppar in en nivå.
2. **Gränserna är satt efter presentation, inte kalender.** En prick per dag i
   ett årfönster skulle kräva 365 prickar som varken går att läsa eller klicka
   på. Därför är `scaleLevel` (kalendisk tidsskala) och `detailLevel` (hur mycket
   en händelse visas) två oberoende koncept - skalan har kalendiska gränser,
   detaljnivån har gränser satta efter hur många markörer som får plats i
   vyhöjden.
3. **Lane-packning löser överlapp.** Händelser och prickar som ligger nära
   varandra i tid flyttas åt sidan i kolumner. Kolumnerna delas bara mellan
   de som faktiskt överlappar (`assignClusteredLanes`), så en ensam händelse får
   hela bredden.
4. **Avståndet mäts i pixlar, inte procent av tidsfönstret**, annars blir
   resultatet olika på mobil och desktop. Därför följs `viewport.clientHeight`
   med en `ResizeObserver`.
5. **Subtilitet som kostade tid:** intervallen i packningen ska utökas med
   kortets *höjd*, inte med avståndet mellan korten. Annars tror algoritmen att
   kort som ligger en timma ifrån varandra överlappar och de delas onödigt
   bredd.
6. **Gotchas:** markörerna grupperas i lokal tid (inte UTC), `intensity` på
   prickarna är `0.45 + 0.55 * min(1, count / 3)`, och antalet visas bara i
   text när det är fler än ett - ett "1" säger ingenting. Ett gemensamt
   `icon`-fält används för både emoji och Iconify-namn.

## Verktyg

- `src/nuxt/app/utils/calendarDetail.ts` - ren logik för nivåer, gruppering och
  packning.
- `src/nuxt/features/calendar-detail.feature` - Cucumber-tester.

## Kopplingar

- [[notes/kalenderns-detaljnivaer-fem-nivaer-och-lane-packning]]
- [[notes/vertikal-kalender-med-ikonstod-i-plan]]
- [[notes/vertikal-kalender-med-kontinuerlig-zoom]]
- [[notes/kalenderns-zoom-styrs-av-vertikal-scrollning]]
