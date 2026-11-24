---
title: Nuxt-CLI via bun run dev går sönder i stk
type: note
tags: [bug, nuxt, bun, projekt:stk]
---

I stk (och repo-template) är bun's .bin/-shim för nuxt trasig: bun kopierar binärfilen istället för att symlänka den, så node_modules/.bin/nuxt importerar '../dist/index.mjs' och Node hittar fel sökväg (node_modules/dist/index.mjs) -> ERR_MODULE_NOT_FOUND. Workaround: 'node node_modules/nuxt/bin/nuxt.mjs dev|build'. Samma fel uppstår i GitHub Actions-workflowen for pages om den använder 'bun run generate' - kontrollera workflowen.
