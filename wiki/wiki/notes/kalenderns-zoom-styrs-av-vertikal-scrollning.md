---
title: Kalenderns zoom styrs av vertikal scrollning
type: note
tags: [feature, kalender, ui, beslut, projekt:plan]
created: 2026-11-17
updated: 2026-11-17
status: active
---

Kalenderdesignen ändrades efter implementation: horisontell zoom tas bort. Vertikal scrollning ska i stället komprimera eller expandera tidslinjen kontinuerligt. Nedåtgående och snabb scrollning visar mer framtid snabbare; uppåtgående/långsam scrollning zoomar in långsammare. Starttiden hålls kvar medan visibleDuration ändras. Komponenten VerticalCalendar använder nu en multiplikativ zoomskala med 2 timmar till 180 dygn.
