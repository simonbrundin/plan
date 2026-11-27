---
title: Kortkommando f: byta målfas (status) för ett eller flera mål
type: note
tags: [feature, goals, tangentbord, status, projekt:plan]
---

ÖNSKEMÅL: ett kortkommando som ändrar ett måls status (målfas) och som fungerar överallt där man kan markera mål - både för ett enskilt mål och för flera samtidigt.

VARFÖR: Målfasen (status_id: definiera/planera/genomföra/verifiera/klar) var den enda egenskapen på ett mål som INTE gick att ändra från tangentbordet - bara via en rullmeny i fas-ikonen. Övriga flaggor (klart, evigt, påbörjat, agent, nu) hade redan tangenter. Nu saknades en enhetlig tangent.

HUR: tangenten 'f' växlar målfasen framåt i cykeln Definiera -> Planera -> Genomföra -> Verifiera -> Klar -> tillbaka till Definiera. 'Shift+F' går bakåt. Samma tangent gäller:
- /goals: det markerade målet, eller alla Tab-markerade mål om något är markerat
- /now: samma logik
- /goal/:id: Tab-markerade undermål, annars förälder i föräldrläge, annars huvudmålet, annars det markerade undermålet
Eviga mål hoppas över (de har ingen fas - ikonen döljs för dem).

IMPLEMENTATION: ny composable src/nuxt/app/composables/useGoalPhase.ts (cyclePhases/setPhase, delar fas-cykeln mellan sidor), rent hjälpfunktion nextPhase() i src/nuxt/app/utils/goalPhase.ts (enhetstestad), tangenterna i goals/index.vue, now.vue och goal/[id].vue.

VARNING: 'f' var ledigt på alla tre sidor, men leader-läget har fortfarande 'f'+'d' (visa/dölj avslutade mål på målsidan) - de krockar inte eftersom ledern kräver Space först.

VERIFIERAT: enhetstester (241 pass), typecheck-gaten grön, build grön, samt end-to-end i headless Chromium mot den lokala dev-instansen: f / f igen / Shift+F på /goals, Tab+Tab+f (bulk: två mål bytte fas, det tredje orördes), f på /goal/:id och f på /now.
