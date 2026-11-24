---
title: Kalendern: en kolumn, dolda serier och en inställningssida
type: note
tags: [kalender, feature, zoom, inställningar, session, projekt:plan]
created: 2026-11-18
updated: 2026-11-18
status: active
---

Användaren ville inte ha rader bredvid varandra. Beslutet blev: ALLA händelser i en enda kolumn, och om det inte ryms så döljs något.

**Varför:** Kolumner delar bredden, så en rad i kolumn 3 av 5 har ~20 % av bredden kvar och titeln klipps till 'Morgon…'. I en kolumn får varje rad hela bredden.

**Så fungerar det nu:**
- En kolumn. Rader som krockar i tid skjuts ned under varandra (tidsaxeln blir ungefärlig, men raden visar sin exakta tid).
- Radens höjd följer händelsens längd, minst 16 px, och en lång händelse begränsas till en fjärdedel av vyn.
- Döljning i trappa: dagliga serier över gränsen (standard 7) -> om kolumnen ändå inte ryms, serien med flest förekomster -> enstaka händelser döljs ALDRIG, de hamnar i '+N till'-knappen med panel.
- Dold serie står kvar i sammanfattningen märkt 'dold' med knapp 'Visa ensam'. Det är viktigt: en dold serie får aldrig försvinna tyst, en daglig rutin kan vara viktig.
- Hysteresis: en dold serie förblir dold tills antalet förekomster faller UNDER gränsen. Annars fladdrar serier vid scroll.
- Gränsen går att ändra på /settings (ny sida, sparad i localStorage, inte i kontot).

**Var:** app/utils/calendarLayout.ts (radpackning), app/utils/calendarVisibility.ts (döljningsplan, ren funktion med hiddenBefore som input för hysteresen), app/composables/useCalendarSettings.ts, app/pages/settings.vue, app/components/VerticalCalendar.vue.

**Lärdom:** En watch med immediate:true i ett setup-script krockar med TDZ om den läser en ref som deklareras längre ner ('Cannot access viewportHeight before initialization'). Flytta koden eller kör planen efter att setup är klar.
