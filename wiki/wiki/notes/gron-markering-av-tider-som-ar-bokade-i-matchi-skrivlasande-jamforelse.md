---
title: Grön markering av tider som är bokade i MATCHi (skrivläsande jämförelse)
type: note
tags: [stk, matchi, bokning, feature, frontend, projekt:stk]
created: 2026-11-06
updated: 2026-11-06
status: active
---

Nytt GET /matchi/bookings i Go API:t (paket internal/occupancy) jämför kommunens bokningar med MATCHis schema och returnerar de bokningar som finns i båda systemen. Endpointen är strikt skrivläsande - den bokar aldrig något i MATCHi (användaren ville uttryckligen inte att något skulle bokas).

Varför: /bookings visar kommunens bokningar, men klubben ville se direkt vilka av dem som faktiskt är bokade i MATCHi (Baldershallen). Grön markning + badge 'Bokad i MATCHi'.

Viktiga lärdomar:
- MATCHi-kontot är själva anläggningen Baldershallen (facility id 531), så schemat från /facility/booking/index gäller redan rätt anläggning - ingen facility-val behövs i klienten.
- Kopplingen kommunresurs <-> MATCHi-bana ligger i court_mappings (628-637 -> '1. Idrottshall'/'1. Squash' etc.). Matchningen sker på resursens namn (location), eftersom kommunens API inte returnerar resources-arrayen.
- Varje schema-dag är en ~1 MB HTML-sida och tar ~5 s. Lösning: hämta bara dagar som har bokningar, parallellt (4 samtidigt) + 90 s cache i occupancy-tjänsten.
- Resurser utan koppling i court_mappings (t.ex. 'Baldershallen Coworking-rum') kan aldrig markeras som bokade i MATCHi.
