---
title: Dashboard: rekommendationsvyn som enda forstavy
type: note
tags: [dashboard, ux, recommendation, projekt:dashboard]
created: 2026-11-26
updated: 2026-11-26
status: active
---

Slutbehovet kristalliserades: 'vilken modell ska jag köra genom vilken provider eller prenumeration'. Ny ModelsRecommendation (beräkning i app/utils/recommendation.ts + tester) som sidans enda default-vy: val av referensnivå (AA Coding Agent-poäng) + agentuppgifter/dag -> SVAR-ruta (billigast: plan eller API), 'utan prenumeration'-ruta och två kompakta listor (med/utan prenumeration). Hållbart strikt: bara AA:s mätta agentkostnader, inga token-estimat i rekommendationen. Plan kvalificerar bara om PLANENS BÄSTA modell klarar nivån (ChatGPT Plus faller bort på Sonnet-nivå eftersom Astra = 53.5 < 56.3). Allt annat (header, filter, stats cards, pareto-front, budget-finder, prenumerations-tabell, full jämförelsetabell, metodiknot) flyttat bakom 'Visa all data'-expander på models-sidan. Navbar-titel 'Kodmodeller - vad ska du köra?'. Verifierat utfall vid 5 uppg/dag på Sonnet-nivå: Copilot Pro /bin/bash.067/uppg (lönar sig vid 1 uppg/mån), utan prenumeration Gemini 4 Neon .86. 39/39 tester, lint, typecheck rent utöver talos, build, Tilt Ready.
