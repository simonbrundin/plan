---
title: Barnordning följer beroenden i målsidan
type: note
tags: [feature, frontend, mål, bugg, projekt:plan]
created: 2026-11-20
updated: 2026-11-20
status: active
related: [notes/sortering-av-undermal-med-piltangenterna-var-trasig-i-plan]
---

Undermål som väntar på ett annat undermål i samma lista ska ligga nedanför det
det väntar på - annars ser det konstigt ut när det indenterade barnet hamnar
ovanför det mål det väntar på.

**Lösning: order-fältet skrivs om, inte vyn.** Listan visar order-fältet (se
[[notes/sortering-av-undermal-med-piltangenterna-var-trasig-i-plan]] - att
sortera i vyn gjorde att flytter inte syntes). Därför skriver
`enforceDependencyOrder()` i `src/nuxt/app/pages/goal/[id].vue` om order-fältet
när ett barn ligger fel: ny util
`src/nuxt/app/utils/childDependencyOrder.ts` (stabil topologisk sortering,
cykler bryts med ursprungsordningen, beroenden till mål utanför listan och till
redan klara mål ignoreras). Ordningen sparas via samma reorder-endpoint som
piltangenterna använder.

Följder:
- Ett barn som väntar på ett annat barn får inte flyttas ovanför det med
  ArrowUp/ArrowDown (flytten avbryts) - annars skulle invarianten brytas direkt.
- `hydrate()` anropar `enforceDependencyOrder()`, så ordningen rättas till när
  sidan laddas och när ett beroende skapas.

**Fallgropar (TDZ):** `hydrate()` kör under setup, så alla `let`-deklarationer
som den kedjan beror på måste ligga före anropet. `pendingOrder`/`isSavingOrder`
flyttades därför till toppen av `<script setup>` - annars kastades
"Cannot access 'ct' before initialization" och ordningen sparades aldrig.

Verifierat i dev-klustret (Tilt, localhost:3100) med [test]-mål: beroende barn
flyttas under sitt beroende, order-fältet skrivs om i databasen, och testdatan
rades bort efteråt.
