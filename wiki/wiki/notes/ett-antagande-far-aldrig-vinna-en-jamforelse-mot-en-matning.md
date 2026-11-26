---
title: Ett antagande får aldrig vinna en jämförelse mot en mätning
type: note
tags: [dashboard, modeller, kvoter, metodik, projekt:dashboard]
created: 2026-11-26
updated: 2026-11-26
status: active
---

Anledning: dashboardens modellsida var "helt oduglig" för att få ut information. Diagnosen
var inte att det saknades data - det saknades ett underlag som gick att läsa.

Tre konkreta problem:
1. Rekommendationen visade "≤ $0,20 per lyckad uppgift" för Google AI Pro, 35x billigare
   än API. Siffran bygger på att kvoten ("4× Pro-nivån") täcker all användning, men den
   kvoten har inget publicerat dollarvärde. Ett antagande presenterades som en mätning.
2. Alternativlistorna visade kostnad men inte poäng - kvalitet gick inte att väga mot pris.
3. Allt användbart (Pareto-front, budgetvy, jämförelsetabell) låg bakom "Visa all data".

Lösningen (commit 6e24cf1):
- Ny beslutstabell överst: modell | poäng | 95 % CI | agent | $ per lyckad uppgift |
  $ per månad | prenumeration. En rad per modell, all vetting-underlag samlat.
- Referensmodellen är valfri. Utan den visas kvalitets-/prisfronten (Pareto) direkt.
- Ny regel: endast alternativ med verifierbar kostnad får vinna. Ett API-pris från en
  mätning är verifierbart. En plan är verifierbar bara om kvoten är publicerad i dollar
  (Copilot credits). Planer med mjuka gränser eller anropskvoter redovisas i en egen lista
  "Kan bli billigare - men bara om kvoten räcker" med brytpunkten mot API-priset, inte
  med ett pris per uppgift.

Den generella regeln att ta med: aldrig jämför ett antagande med en mätning. Om
underlaget inte räcker till ett pris, redovisa brytpunkten i stället.

Kod: app/utils/recommendation.ts (costIsVerifiable, conditionalPlanOptions, paretoFrontier),
app/components/models/ModelsDecisionTable.vue (ny), app/components/models/
ModelsRecommendation.vue, app/composables/useAgentBenchmark.ts (gemensamt benchmark- och
volymstate), 72 tester gröna.
