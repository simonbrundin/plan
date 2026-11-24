---
title: Målgraph: navigering mellan målsidor går från sekunder till millisekunder
type: note
tags: [performance, goals, go, nuxt, session, projekt:plan]
---

PROBLEM: Att byta målsida i plan tog 2-7 sekunder. Målsidan anropade GET /api/v1/goals/:id, och handlern gjorde N+1-frågor: en fråga per mål, per relation, per beroende, per platskoppling. Ett mål med 28 undermål = 36 rundturer till Neon (~40 ms per rundtur, dev-databasen ligger i eu-central-1).

VARFÖR detta var så dyrt: Neon har en rundtur per fråga, så latensen skalar med antalet frågor, inte med datamängden. Databasen var aldrig problemet - antalet rundturer var det.

LÖSNING (två lager):
1. Server: ny GET /api/v1/goals/graph + GetGoal byggs om från N+1 till 4 frågor (mål, relationer, beroenden, platser) som körs parallellt. 3,35 s -> 0,31 s i snitt.
2. Klient: grafen laddas en gång (i layouten vid start) och målsidan byggs ur den i app/utils/goalGraph.ts. Att byta mål = 0 API-anrop, 38 ms (mätt i headless Chromium).

VAD SOM INTE ÄR EN DUM IDÉ: att hämta alla mål direkt från start. Det var redan nästan så - /goals/:id skickade med allGoals varje gång. Det som gjorde det långsamt var relationerna, inte målen. Men nu är det faktiskt en enda fråga i stället för 36.

FALLGROP: app/utils/goalGraph.ts måste ge exakt samma svar som servern. Det verifierades med ett skript som jämför klientens grafen mot serverns svar för alla 144 mål - och det ligger nu som unit test (test/unit/goalGraph.test.ts, körs med 'bun test', inte vitest - vitest finns inte i package.json).

MÄTNING: dev-databasen (Neon, eu-central-1) har ~40 ms per rundtur. Varje extra fråga i ett handlarflöde kostar alltså 40 ms - räkna rundturer, inte frågor.

ÖVRIGT: GET /goals/:id svarar nu 404 (i stället för 403) för mål som tillhör en annan användare. Tidsstämplar i föräldralistan har nu rätt has_parent och lokal tidszon (var Z/UTC + fel has_parent före).
