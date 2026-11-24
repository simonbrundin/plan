---
title: Jobb ska avslutas när körningen är klar, och en delad proxy rensar TypeScript-felen
type: note
tags: [plan, agent, jobb, worker, typescript, refaktor, nuxt, lärdom, projekt:simon]
---

Två saker som blev bättre under slutverifieringen 2026-11-24:

1. Jobbet levde en hel timme efter att körningen var klar. Workern avslutades
   bara när målet nådde verifieringsfasen, men en livsprofil lämnar målet i
   planfasen - så jobbet stod och väntade till deadline (3600 s) utan att
   göra något. Nu avslutas jobbet så snart körningen är klar: jobbet för
   AgentRun #71 tog 6 sekunder i stället för en timme.

   Lärdom: "avsluta när arbetet är gjort" ska gälla det som faktiskt blev
   gjort, inte ett indirekt tecken på att det är gjort (målets fas). En
   profil som inte får göra hela kedjan når aldrig den sista fasen.

2. Nuxt: varje composable hade sin egen apiProxy med samma $fetch-anrop.
   Den typade $fetch jämför varje anrop mot varje route i appen - när
   agentens routes växte blev jämförelsen så djup att TypeScript gav upp
   ("excessive stack depth", TS2321). Felet flyttade sig från fil till fil
   när jag fixade det en i taget: useAgentApi -> useGoalApi -> usePlaces ->
   useTags.

   Lösningen var inte att fixa varje fil utan att lösa dubbleringen. En
   delad proxy (app/utils/apiProxy.ts) med $fetch castad till en enkel
   funktion gör att den typade route-unionen aldrig jämförs mot dynamiska
   sökvägar. Typecheck-felen i agentkoden försvann och det totala antalet
   föll från 344 till 299.

   Lärdom: när ett typefel flyttar sig när du fixar det, är rotorsaken
   inte felet utan arkitekturen runt. Ändra på ett ställe i stället för
   där felet råkar vara.
