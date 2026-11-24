---
title: Taggar via leader+t på alla målsidor
type: note
tags: [feature, frontend, plan, projekt:plan]
created: 2026-11-23
updated: 2026-11-23
status: active
---

Kommandot för att sätta taggar har flyttat från tangenten 'g' på målsidan till leader+t, och fungerar nu på alla sidor där ett mål är markerat (/goals, /now, /goal/:id).

Varför: taggar sätts på det markerade målet, och det finns markerade mål på flera sidor - det var onödigt att kommandot bara fanns på målsidan.

Hur: en ny delad composable useGoalTagPicker (src/nuxt/app/composables/useGoalTagPicker.ts) håller state utanför funktionen (som useGoalDatesModal), så att GoalTagPicker kan renderas i layouten (src/nuxt/app/layouts/default.vue) och öppnas från valfri sida via openFor(goal). Sidor uppdaterar sitt eget urval genom att lyssna på savedRevision och läsa lastSaved.

Följd: 'g' på målsidan är nu ledig igen, och hjälpmodalen (HelpModal.vue) samt AGENTS.md är uppdaterade.
