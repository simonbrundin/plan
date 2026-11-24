---
title: Nya mål får ikon från titeln
type: note
tags: [feature, nuxt, icon, goals, projekt:plan]
created: 2026-11-17
updated: 2026-11-17
status: active
related: [notes/automatisk-malikon-behovde-sokord-och-svenska-alias, notes/malikoner-anvander-enkel-oversattning-fore-iconify]
---

När ett mål skapas via Nuxts useGoalApi söks Iconify med målens titel via /api/icons. Första träffen skickas med i POST /api/goals som icon. Om sökningen misslyckas eller titeln är kort skapas målet ändå och Go-API:t använder sin vanliga heroicons:star-fallback. Detta är centraliserat i useGoalApi så grundmål, föräldramål och undermål använder samma beteende.

## Relaterade

- [[notes/malikoner-anvander-enkel-oversattning-fore-iconify]]
- [[notes/automatisk-malikon-behovde-sokord-och-svenska-alias]]
