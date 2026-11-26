---
title: Kvotmodell: anropskvoter och mjuka gränser kan inte räknas till dollar
type: note
tags: [dashboard, prenumeration, kvot, minmax, projekt:dashboard]
created: 2026-11-26
updated: 2026-11-26
status: active
---

MiniMax Token Plan och M Plan har kvoter i anrop per 5-timmarsfönster (1 500 anrop för Starter) respektive relativa enheter (Go 1x, Explore 3x, Build 7,5x). Inte i dollar. Antagandet att kvotens dollarvärde = planpriset (1:1) var fel och togs bort.

Ny modell med tre ärliga kvottyper:
- credits-usd: kvoten har publicerat dollarvärde (Copilot AI credits, 1 credit = $0.01). Exact kostnad per uppgift = kvotvärde / uppmätt agentkostnad.
- quota-units: kvoten räknas i anrop eller relativa enheter (MiniMax). Antalet täckta agentuppgifter går inte att räkna ut - en agentuppgift består av många anrop. Kostnaden per uppgift är en undre gräns.
- soft-limits: gränserna publiceras inte i dollar (Claude, Google AI, ChatGPT). Bara break-even går att säga, och det gäller bara så länge gränserna räcker.

Viktigare ändring: bara optioner med uppmätt kodnivå (AA Coding Agent Index) får vinna rekommendationen. En plan med okänd kodnivå kan se billig ut (spridningskostnad $10/150 = $0.07/task) och ändå vara för svag. Detta filtreras i buildRecommendation via qualityKnown, och planOptions sorteras så att uppmätta nivåer hamnar först.

MiniMax-planernas modellslugs korrigerade: Token Plan innehåller M2.7, M Plan innehåller M3.1 Flash (inte alla modeller i alla planer). Arspriser tillagda: Go $220/år, Explore $550/år, Build $1320/år.

Verifierat: 52/52 tester, riktad ESLint grön, build grön, /models HTTP 200.
