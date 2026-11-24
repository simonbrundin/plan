---
title: Mellanslag blockerades i målbeskrivningen
type: note
tags: [bug, goals, nuxt, keyboard, projekt:plan]
created: 2026-11-17
updated: 2026-11-19
status: active
related: [notes/leader-key-gjordes-global-mellanslag-sekvenser-pa-alla-sidor]
---

Global navigation i src/nuxt/app/composables/useGlobalNavigation.ts hanterade event.key === ' ' och körde preventDefault innan isInputFocused kontrollerades. Därför gick mellanslag inte att skriva i textarea för målbeskrivningar. Flytta input/textarea-kontrollen före leader-mode-logiken så att mellanslag och andra vanliga tecken lämnas till fältet. Verifierat med bun test (27 pass) och bun run build.

## Relaterade

- [[notes/leader-key-gjordes-global-mellanslag-sekvenser-pa-alla-sidor]]

