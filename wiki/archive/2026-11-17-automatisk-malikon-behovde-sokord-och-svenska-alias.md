---
title: Automatisk målikon behövde sökord och svenska alias
type: note
tags: [bug, feature, icon, goals, projekt:plan]
---

Första implementationen sökte Iconify med hela målens titel. Iconify gav ofta inga träffar för svenska titlar eller hela fraser, vilket gjorde att Go-API:ts heroicons:star-fallback användes för alla nya mål. useGoalApi söker nu först relevanta svenska alias (t.ex. träna/träning→fitness, bok/läsa→book, hem→home, ekonomi→money) och därefter enskilda ord innan fallback.
