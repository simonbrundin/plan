---
title: Köda körningar är uttryckliga begäran - och en avslutad körning har alltid run_completed
type: note
tags: [plan, agent, jobb, worker, bugg, händelser, verifiering, projekt:simon]
---

Två brister hittades när flera mål kördes parallellt i dev (AgentRun #65, #66, #67, #68, #69):

1. Workern hämtade bara mål markerade som agentmål (GET /api/v1/goals?agent=true).
   En köad körning för ett annat mål blev liggande i Plan medan jobbet köade
   i evighet. Fix: när en körning står i kö hämtas målet direkt
   (GET /api/v1/goals/:id) och körningen tas över oavsett agent-flaggan.
   Test: TestKöadKörningTasÖverÄvenFörMålSomInteÄrAgentmål.

2. En körning som avslutades efter definitionen (utan coder) registrerade
   inte run_completed, och en körning som avslutades av profilen registrerade
   inte run_started. Historiken såg då ut att en körning bara upphörde.
   Nu börjar och slutar varje körning med run_started/run_completed.

Lärdom: "en köad körning är en uttrycklig begäran" är en policy som måste
följas i kod, inte bara i dokumentation. En begäran som ignoreras i ett
lager blir en tyst förlust - jobbet går inte att felsöka eftersom det bara
ser ut att vänta.

Parallellkörning fungerar: mål #540 (coding) och #543 (life) hade jobb
samtidigt, ett per mål. Ett mål får ha högst ett pågående jobb - två
körningar mot samma repo skulle skriva över varandra.
