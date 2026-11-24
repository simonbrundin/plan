---
title: Kalenderns scroll-axlar bytta: lodrät zoomar, vågrätt panorerar
type: note
tags: [kalender, feature, interaktion, projekt:plan]
created: 2026-11-17
updated: 2026-11-17
status: active
---

Användaren bad om att byta ordning på scrollaxlarna i den vertikala kalendern (VerticalCalendar.vue): lodrät skrollning ska zooma tiden och horisontell skrollning ska ändra starttiden. Tangentbordet följer samma konvention: j/ArrowDown och k/ArrowUp zoomar (ner zoomar ut, upp zoomar in), h flyttar tiden bakåt och l framåt. Plus/minus zoomar oförändrat.

Varför: den vertikala kalendern visar tiden lodrätt, men användaren ville ha en konvention där scrollaxeln som följer tidsaxeln (lodrät) styr zoom och den tvärgående axeln styr förflyttning - samma konvention som i många tidslinjeverktyg. Drag med pekaren byttes på samma sätt så att alla tre inmatningssätt (scroll, drag, tangentbord) beter sig likadant.

Följd: specs/vertical-calendar.md och hjälpdialogen (HelpModal.vue) uppdaterades så att de beskriver den nya konventionen.
