---
title: Plan API och agent delar Go-modul men behåller tjänstegränsen
type: note
tags: [arkitektur, go, agent, projekt:plan]
created: 2026-11-26
updated: 2026-11-26
status: active
---

src/api och src/agent flyttades till en gemensam Go-modul med rot i src/ (modulnamn plan). Paketprefixen är plan/api/... respektive plan/agent/.... Det minskar friktionen att hantera två Go-moduler och låter dem dela dependency graph, men binärerna, Docker-images, deploymenter, databaser och HTTP-kontraktet förblir separata. Skälet är att agenten behåller egen exekverings-/säkerhetsgräns och att denna refaktor inte ska slå ihop deras runtime eller dataägande.
