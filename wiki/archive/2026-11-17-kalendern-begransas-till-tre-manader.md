---
title: Kalendern begränsas till tre månader
type: note
tags: [kalender, feature, zoom, projekt:plan]
---

VerticalCalendar visar nu maximalt 90 dygn åt gången. När fortsatt nedåtgående zoomning passerar gränsen behålls tidsfönstret på tre månader och startTime flyttas framåt med överskjutande intervall. Ändringen finns i src/nuxt/app/components/VerticalCalendar.vue och specs/vertical-calendar.md.
