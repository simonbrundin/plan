---
title: Kalendern: sammanfattningen av serier togs bort
type: note
tags: [kalender, beslut, ui, projekt:plan]
---

Användaren ville inte se de 'blobbar' (piller) överst i kalendern som visade antalet förekomster per återkommande serie. Vi tog bort hela sammanfattningsraden (.recurrence-summary) i src/nuxt/app/components/VerticalCalendar.vue, plus notisen 'Visa bara X' (focus-notice).

Följder: 'Visa ensam' och 'hoppa till nästa förekomst' försvann, så focusKey togs bort ur calendarVisibility.ts (planVisibility), enhetstestet och feature-scenariot 'En serie kan visas ensam'. Serierna räknas fortfarande (recurringSeries) - bara för döljningsreglerna.

Varför: vyn ska vara ren, användaren ville ha färre saker ovanför tidslinjen.

Konsekvens att känna till: döljningen av återkommande serier är nu tyst - en dold serie syns inte i tidslinjen och syns inte någon annanstans. Zooma in på serien för att se den. Specen specs/calendar-event-rows.md är uppdaterad, liksom texts i settings.vue och useCalendarSettings.ts.
