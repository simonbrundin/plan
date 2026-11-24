---
title: Föräldrar syns på mål i /now
type: note
tags: [feature, now, goals, projekt:plan]
created: 2026-11-22
updated: 2026-11-22
status: active
---

Målsidan /now (Gör nu) visar nu vilka föräldrar ett mål ligger under.

Varför: listan visade bara målets egen titel, så det gick inte att se sammanhanget utan att öppna varje mål.

Hur: GET /api/v1/goals/prioritized returnerar nu fältet parents (id + titel för alla föräldrar, sorterat på relationsvikten). Tidigare fanns bara parentTitle/parentId, som pekade ut föräldern med högst vikt. Aggregeringen sker i SQL (json_agg) i GetPrioritizedGoals, och JSON-listan tolkas i parseGoalParents.

Klienten (usePriorityMode/now.vue) renderar föräldrarna som små klickbara länkar i raden, separerade med ·, med tooltip listing alla föräldrar.
