---
title: Viktning av barnmål med balanserad draglinje
type: note
tags: [feature, goals, frontend, api, projekt:plan]
---

Plan har fått första implementationen av viktning mellan förälder och barnmål. UI:t visar en mittlinje för lika vikt och låter användaren dra ett mål åt vänster/höger; resten fördelas proportionellt och sparas atomiskt via PATCH /api/goals/relations/weights. Viktfördelningen representeras som heltalsprocent som summerar till 100 i den nya bulk-endpointen. Nya filer: src/nuxt/app/components/GoalWeightEditor.vue och utils/goalWeights.ts. Nuxt build och Go-test passerar. Kvar att ta ställning till: hur nya/borttagna barnmål automatiskt ska fördela vikt samt eventuell migrering av äldre relativa weight-värden.
