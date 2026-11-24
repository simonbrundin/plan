---
title: p-tangenten konfliktade med viktläget på /priority
type: note
tags: [bug, frontend, kortkommandon, projekt:plan]
created: 2026-11-20
updated: 2026-11-20
status: active
---

Fynd: usePriorityToggle växlade prioritetssidan vid varje tryck på 'p', samtidigt som usePriorityMode försökte aktivera viktläget med samma tangent. Resultat: viktläget (P+J/K för att ändra vikt) gick aldrig att använda, eftersom 'p' först slog till mode-flaggan och sedan navigerade bort från /priority.

Fix (2026-11-20): viktlädet aktiveras nu med Shift+P (håll ned), medan vanligt 'p' fortsätter att växla mellan prioritetssidan och föregående sida. usePriorityMode håller nu klistan på rå event.key ('P' = shift+p) och usePriorityToggle ignorerar tangenttryck med shift nedtryckt.

Varför shift: layoutens indikator och prioritetssidans hjältext säger 'P+K/J' - versal P är naturligt läsbart som shift+p, och ingen befintlig funktion använder shift+p tidigare.

Verifierat headless (byggd app + mockat API-svar): shift+p stannar på /priority, indikatorn 'Prioriteringsläge' visas, K höjer vikten 10->11, och läget avslutas när tangenten släpps. Vanligt 'p' navigerar fortfarande bort från /priority.
