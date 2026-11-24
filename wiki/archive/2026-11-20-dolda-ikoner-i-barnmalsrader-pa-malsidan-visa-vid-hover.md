---
title: Dolda ikoner i barnmålsrader på målsidan (visa vid hover)
type: note
tags: [ui, ikon, målsida, projekt:plan]
---

Varje barnmål på målsidan (src/nuxt/app/pages/goal/[id].vue) hade 8 ikoner i raden, även om målet inte använde agent, plats, deadline, upprepning eller tidsåtgång.

Beslut: ikonerna för de funktionerna visas bara om barnmålet faktiskt använder dem. Funktionen i sig finns kvar - tangenterna (c, r, p, a, m) och musknapparna fungerar - men en oanvänd ikon döljs i vila och visas igen när raden svävas över (Tailwind group-hover) eller får fokus (focus-within). Wrapper-spanet får 'pointer-events-none' i vila så en osynlig knapp inte går att klicka av misstag.

Implementation: optionalIconClass(active) i goal/[id].vue. AgentBadge, deadline, EstimateBadge, upprepning och platser omsluts av en span med den klassen. Målets egen ikon, start/paus (s), beroende-draghantag och etiketten visas alltid.

Varför hover och inte v-if: annars skulle man inte kunna lägga till en plats/deadline på ett barnmål med musen, bara via tangent. Det skulle vara en funktionsförlust, vilket användaren uttryckligen inte ville.

Verifierat headless (playwright + chromium, lokal Nuxt-build mot Go-API på port 8099 som pekar mot Neon dev): utan hover har alla fem wrappers opacity 0/pointer-events none, med hover opacity 1/auto, och barnmål som har agent eller tidsåtgång visar de ikonerna utan hover.
