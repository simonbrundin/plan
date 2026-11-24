---
title: Kalenderns tidsaxelrader animeras vid zoom
type: note
tags: [kalender, animation, zoom, projekt:plan]
---

Tidsaxelns tick-rader och månadsgränser renderas i egna Vue TransitionGroup-lager i VerticalCalendar.vue. När zoomningen byter tickintervall eller månadsmarkörer passerar ettårsgränsen får element enter/leave-animation med duration från --detail-duration, som följer scrollhastigheten. Positionerade rader använder CSS-egenskapen translate i stället för transform så att Vue:s FLIP-move kan interpolera förflyttningen utan att skriva över den vertikala centreringen.
