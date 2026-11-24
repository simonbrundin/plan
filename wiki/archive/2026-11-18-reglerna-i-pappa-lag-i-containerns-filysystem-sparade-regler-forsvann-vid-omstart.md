---
title: Reglerna i pappa låg i containerns filysystem - sparade regler försvann vid omstart
type: note
tags: [bug, mail, pappa, kubernetes, tilt, projekt:pappa]
---

Symtom: I /mail/rules (pappa) såg det ut som att "Spara regler" inte funkade - samma regler syntes igen efter en uppdatering av sidan.

Varför det hände (2026-11-18):
1. Reglerna lagrades i mailagentens containerns filsystem (/app/rules.md). Varje gång containern/poden skapades om (t.ex. när Tilt byggde om imagen) resetades filen till den version som bakades in i imagen - den redigerade versionen var alltså borta, även om PUT:en hadde svarat "ok".
2. Två tilt-sessioner kördes parallellt mot samma projekt (~/repos/pappa/environments/dev). Den ena hade en stale fil-watch och byggde en gammal image - deploymenten fick args för den nya binaryn men en gammal image, och mailagenten gick i CrashLoop med "flag provided but not defined: -rules-default".
3. Frontendens save() trodde att ett "ok" betydde att regeln sparades, och navigerade bort från sidan efter 1 sekund - felmeddelandet syntes aldrig.

Fix (2026-11-18):
- Reglerna ligger nu på en PVC (environments/base/mailagent-rules-pvc.yaml), monterad på /app/rules. -rules pekar på /app/rules/rules.md.
- main.go har flaggan -rules-default; rules.SeedFile() fyller volymen med den inbakade versionen första gången (annars skulle appen inte ens starta). Funktionen har egna tester i src/go/rules/seed_test.go.
- GET /api/v1/rules svarar nu med Cache-Control: no-store, så ingen cache kan visa en gammal version.
- Frontendens save() läser tillbaka det som ligger kvar på servern efter PUT:en och jämför - en falsk "ok" ger ett fel i stället för en bekräftelse. Felmeddelanden innehåller HTTP-status och URL (describeApiError i src/nuxt/shared/utils/api-error.ts, testad med bun).
- Sidan navigerar inte bort automatiskt efter sparning - resultatet ska synas.

Verifierat: sparning via webbläsaren (Playwright) → omstart av mailagent-poden → regeln ligger kvar. Det fungerade INTE innan fixen.

Lärdom: "sparas" i ett kluster betyder "ligger på en volym", inte "ligger i containerns filsystem". Kontrollera med kubectl delete pod så att det verkligen överlever.
