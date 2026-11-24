---
title: Mellanslag blockerades i målbeskrivningen
type: note
tags: [bug, goals, nuxt, keyboard, projekt:plan]
---

Global navigation i src/nuxt/app/composables/useGlobalNavigation.ts hanterade event.key === ' ' och körde preventDefault innan isInputFocused kontrollerades. Därför gick mellanslag inte att skriva i textarea för målbeskrivningar. Flytta input/textarea-kontrollen före leader-mode-logiken så att mellanslag och andra vanliga tecken lämnas till fältet. Verifierat med bun test (27 pass) och bun run build.
