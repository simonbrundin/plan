---
title: Undo-stapel i plan: Z ångrar senaste målhandling
type: note
tags: [feature, undo, tangentbord, frontend, projekt:plan]
created: 2026-11-20
updated: 2026-11-20
status: active
related: [notes/tidsatgang-pa-mal-i-plan-estimated-minutes]
---

På begäran: tangenten Z ångrar nu den senaste handlingen i plan (slutfört/pågående/evigt/deadline/upprepning på mål).

Varför: tangenttrycken (d/s/e/c/r) ändrade mål direkt utan återställningsmöjlighet - ett misstag gick inte att backa.

Hur det löses: global undo-stapel i app/composables/useUndoStack.ts (modulnivå, appen körs utan SSR), ren logik i app/utils/undoStack.ts (push/pop, max 25 poster), målens handlingar via app/composables/useGoalUndo.ts som sparar det gamla värdet. Tangenten hanteras i layouts/default.vue (global) och ignoreras i textfält/sök/leader-läge. Efter ångrandet räknas revision upp så att sidor med egna listor (/goals, /priority) hämtar om. Ingen backendändring behövs - API:t har redan anropen (PATCH goals/:id, DELETE goals/:id/completions).

Verifiering: enhetstester (bun test) + end-to-end i headless Chromium mot Go API:t på Neon dev: skapa [test]-mål, tryck d (målet slutfördes), tryck z (finished = null igen), toast 'Ångrade' synlig. Testmålet raderades efteråt - dev-databasen innehåller användarens riktiga data.

Begränsning: går inte att ångra borttagning av mål/händelse eller titel-/beskrivningsredigering. Ångrar man flera förekomster av samma återkommande mål tar API:t bort den senast slutförda.

## Relaterade

- [[notes/tidsatgang-pa-mal-i-plan-estimated-minutes]]

