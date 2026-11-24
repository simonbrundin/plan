---
title: Dold skrivbordsmeny med leader-kommando
type: note
tags: [frontend, navigation, session, projekt:plan]
created: 2026-11-20
updated: 2026-11-20
status: active
---

Plan: skrivbordsmenyn använder nu samma slideover-beteende som mobilmenyn och är dold tills den öppnas med menyikonen. Delat state i src/nuxt/app/composables/useSidebarState.ts gör att leader-kommandot G+M (leader + M) växlar menyn. Verifierat med bun run build i src/nuxt.
