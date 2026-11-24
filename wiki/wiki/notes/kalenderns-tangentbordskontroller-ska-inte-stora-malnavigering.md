---
title: Kalenderns tangentbordskontroller ska inte störa målnavigering
type: note
tags: [frontend, bug, projekt:plan]
created: 2026-11-20
updated: 2026-11-20
status: active
---

VerticalCalendar ligger permanent i default-layouten och har därför en global keydown-lyssnare. Den måste ignorera j/k/h/l när användaren är på målsidor; annars zoomar/flyttar kalendern samtidigt som useGlobalNavigation navigerar mål. Fixen begränsar kalenderns globala tangentkontroller till /calendar. Vid verifiering mot lokal dev-login svarade /api/proxy/calendar-events?scope=all med [] för dev-användaren, så uteblivna event där beror på att den användaren saknar eventdata och inte på layout-renderingen.
