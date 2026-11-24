---
title: Kalenderns händelser låg över modaler och Escape var dövt
type: note
tags: [bug, frontend, z-index, kalender, projekt:plan]
---

Två buggar hittade när kortkommando-hjälpen (HelpModal) skulle visas ovanpå kalendern 2026-11-20.

1) z-index: händelse-chipsen i VerticalCalendar får z-index = sin pixelposition i tidslinjen ('z-index': Math.round(row.top)), alltså värden som kan vara tusentals. Eftersom chipsen låg i root-stacking-contexten hamnade de OVANPÅ alla modaler (z-index 100). Fix: .calendar-shell fick 'isolation: isolate', så kalendern blir en egen staplingskontext och dess interna z-index inte kan nå ut i root-contexten. Händelsedialogen (event-modal-backdrop, z-index 100001) ligger i en Teleport till body och påverkas inte.

2) Escape: GoalDatesModal lyssnar på window keydown I CAPTURE-fasen och sväljde Escape (preventDefault + stopPropagation) även när popupen var stängd. Eftersom capture-fasen körs före bubble-fasen nådde Escape aldrig till andra lyssnare - varken HelpModal, snabbfältet eller kalenderns egen dialogruta kunde stängas med Escape. Fix: lyssnaren returnerar direkt om props.open är falskt. Viktigt att lyssnaren är kvar i capture-fasen - annars skulle Escape nå sidans egna tangentbevakare först.

3) Kalenderns egen händelsedialog kunde inte stängas med Escape: handleCalendarKeydown ignorerade tangenttryck när fokus låg i ett INPUT, och dialogen autofokuserar titelfältet. Fix: Escape hanteras före input-kollen i handleCalendarKeydown.

Verifierat headless (byggd app, mockat API-svar): elementFromPoint/elementsFromPoint visar att modalen ligger över händelsen med isolation (utan isolation ligger händelsen överst); Escape stänger hjalpmodalen, händelsedialogen, snabbfaltet och leader-modalen.
