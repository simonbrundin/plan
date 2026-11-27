---
title: Kortkommando s: byta målfas (status) och bulkmarkera klart för ett eller flera mål
type: note
tags: [feature, goals, tangentbord, status, projekt:plan]
created: 2026-11-27
updated: 2026-11-27
status: active
related: [concepts/anvanda-agenter-i-plan]
---

ÖNSKEMÅL: ett kortkommando som ändrar ett måls status (målfas) och som fungerar överallt där man kan markera mål - både för ett enskilt mål och för flera samtidigt.

VARFÖR: Målfasen (status_id: definiera/planera/genomföra/verifiera/klar) var den enda egenskapen på ett mål som INTE gick att ändra från tangentbordet - bara via en rullmeny i fas-ikonen. Övriga flaggor (klart, evigt, påbörjat, agent, nu) hade redan tangenter. Nu saknades en enhetlig tangent.

HUR: tangenten 's' växlar målfasen framåt i cykeln Definiera -> Planera -> Genomföra -> Verifiera -> Klar -> tillbaka till Definiera. 'Shift+S' går bakåt. Samma tangent gäller:
- /goals: det markerade målet, eller alla Tab-markerade mål om något är markerat
- /now: samma logik
- /goal/:id: Tab-markerade undermål, annars förälder i föräldrläge, annars huvudmålet, annars det markerade undermålet
Eviga mål har också en fas - de kan gå från Definiera till Planera och vidare. Bara klarmarkeringen (bulk-'d') gäller dem inte: eviga mål kan inte slutföras.

TANGENTEN ÄNDRAD: första versionen använde 'f'/'Shift+F'. Användaren valde 's'/'Shift+S' eftersom 'f' kunde misstas för något annat. 's' är ledig utan leader på alla tre sidor - 'Space'+'s' (påbörjat) och 'Space'+'f'+'d' (visa/dölj avslutade mål) krockar inte eftersom ledern kräver Space först.

BULK KLART ('d'): tangenten 'd' (slutförd) tog tidigare bara hand om ett mål i taget. Nu gäller den hela Tab-urvalet: om alla valda mål redan är klara tas klarmarkeringen bort, annars sätts den på - samma princip som övriga bulkändringar (nu/påbörjat/agent/evigt). Eviga mål hoppas över. Återkommande mål slutförs en förekomst i taget (via completeGoalOccurrence) och kan inte "öppnas" igen - så de hoppas över när klarmarkeringen ska tas bort. Detta gäller på /goals, /now och /goal/:id. Bulkverktyget har fått en "Klar"-knapp som går via samma funktion.

IMPLEMENTATION: ny composable src/nuxt/app/composables/useGoalPhase.ts (cyclePhases/setPhase, delar fas-cykeln mellan sidor), rent hjälpfunktion nextPhase() i src/nuxt/app/utils/goalPhase.ts (enhetstestad), tangenterna i goals/index.vue, now.vue och goal/[id].vue. Bulk-klart: toggleBulkFinished() per sida + emit-typen i GoalBulkActions.vue ('finished').

VERIFIERAT: enhetstester (241 pass), typecheck-gaten grön, build grön, samt end-to-end i headless Chromium mot den lokala dev-instansen: 's' / 's' igen / 'Shift+S' på /goals, Tab+Tab+'s' (bulk: två mål bytte fas, det tredje orördes), 's' på /goal/:id och på /now, samt bulk-'d' (två mål klara, det tredje orört) på /goals, /now och /goal/:id.

OBS: klara mål visas inte i listorna (/goals filtrerar bort slutförda mål), så efter en bulk-markering försvinner de urvalda målen ur vyn - det är väntat beteende, inte en bugg.

## Relaterade

- [[concepts/anvanda-agenter-i-plan]]

