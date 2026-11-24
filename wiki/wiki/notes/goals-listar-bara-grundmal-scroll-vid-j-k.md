---
title: /goals listar bara grundmål + scroll vid j/k
type: note
tags: [feature, fix, goals, projekt:plan]
created: 2026-11-16
updated: 2026-11-16
status: active
---

Ändring i plan-repot (commit 36b89d8, 9ae273a, 9be938f).

Varför: /goals blandade grundmål och undermål, så samma mål syntes på flera ställen. Nu returnerar Go-API:t has_parent (EXISTS-subquery mot goal_relations) och Nuxt filtrerar med isRootGoal() i utils/goalFilters.ts - filterlogiken är testad (test/unit/goalFilters.test.ts).

Fix: /goals saknade scrollIntoView-watch på selectedIndex, så j/k flyttade markeringen men inte vyn. Löst med data-goal-index + watcher, samma mönster som goal/[id].vue.

Notera: remote svarade 'Changes must be made through a pull request' men pushen gick igenom till main.
