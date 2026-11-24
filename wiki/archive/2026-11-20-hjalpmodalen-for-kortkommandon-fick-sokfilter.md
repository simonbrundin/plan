---
title: Hjälpmodalen för kortkommandon fick sökfilter
type: note
tags: [frontend, kortkommandon, helpmodal, projekt:plan]
---

2026-11-20: HelpModal.vue fick ett sökfilter. När modalen öppnas får fältet fokus automatiskt (watch på props.open + nextTick + focus), så man kan börja filtrera direkt. Filtret matchar både tangentnamn och beskrivning, och flera ord kräver alla stämma (AND). Sektioner utan träff döljs helt, en tom-text visas annars, och en räknare visar 'X av Y kommandon matchar'.

Sökfältet ligger i ett formulärfält, så sidornas tangentbevakare (målsidan, /goals, /priority, kalendern, layoutens globala handler) ignorerar tangenttrycken - de har alla en isInInput-guard. Undantag att hålla koll på: ctrl+k öppnar fortfarande appens globalsök, och Escape stänger modalen.

AGENTS.md i plan-repot uppdaterades med att modalen är data-driven - nya tangenter läggs in i 'sections' i script-blocket, och filtret filtrerar utifrån samma data.
