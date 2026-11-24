---
title: Åtta feature-commits pushade till det arkiverade mailmaster-repot
type: note
tags: [mailmaster, arkiv, push, kameo, oauth, projekt:mailmaster]
created: 2026-11-20
updated: 2026-11-20
status: active
---

Arbetet låg som ocommittade ändringar i arbetsytan i det arkiverade mailmaster-repot och skickades i 8 commits (regelparsering, AI-svar med flera åtgärder, actions-executor, OAuth-paket, MCP-sessions/analys-API, Nuxt-kontoval + analys-vy, Kameo-utbetalningssida, SOPS-synk av Kameo-uppgifter) till origin/main.

Varför: motsvarande kod finns redan, och är den levande versionen, i ~/repos/pappa - bl.a. src/go/mcp/{analysis,auth,session}.go, src/go/oauth/ och sidan /kameo-utbetalningar. Arkivrepot är alltså en eftersläpande spegel.

Gäller framöver: mailmaster-repot är arkiv - ändra inte där. Vill man ändra mailagenten eller Kameo-flödet ska det göras i pappa. Pushar till arkivet är spegelhistoria, inte en ny utvecklingsgren.
