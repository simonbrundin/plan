---
title: Ikonsökning får timeout och checkbox-fallback vid nya mål
type: note
tags: [bug, icon, goals, nuxt, projekt:plan]
---

När ett nytt mål skapades väntade useGoalApi obegränsat på /api/icons, exempelvis om översättningstjänsten/MiniMax inte svarade. Ikonuppslagningen har nu 5 sekunders timeout och returnerar lucide:check-square vid fel, så POST /api/goals ändå genomförs med en synlig checkbox som fallback.
