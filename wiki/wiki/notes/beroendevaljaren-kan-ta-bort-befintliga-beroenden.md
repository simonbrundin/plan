---
title: Beroendeväljaren kan ta bort befintliga beroenden
type: note
tags: [feature, keyboard, dependencies, projekt:plan]
created: 2026-11-22
updated: 2026-11-22
status: active
---

Uppföljning på beroendeväljaren: när användaren trycker w visas nu befintliga beroenden överst i GoalDependencyPicker.vue. Varje befintligt beroende har en x-knapp som anropar befintlig DELETE /api/v1/goals/dependencies och laddar om grafen. Detta fungerar från både målsidan och mållistan.
