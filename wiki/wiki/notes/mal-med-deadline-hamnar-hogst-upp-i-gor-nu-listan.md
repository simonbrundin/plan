---
title: Mål med deadline hamnar högst upp i Gör nu-listan
type: note
tags: [feature, plan, nuxt, prioritetsida, projekt:plan]
created: 2026-11-28
updated: 2026-11-28
status: active
---

# Mål med deadline hamnar högst upp i Gör nu-listan (2026-11-28)

## Vad

I `/now` finns en toggle **"Deadline först"**. När den är på hamnar mål med
deadline överst i listan, den närmaste fristen först. När den är av gäller den
manuella ordningen (priority_order, vikt, skapad) som tidigare.

Toggle:en är **av som standard** - den manuella ordningen är listans huvudregel,
deadline-först är ett val för den som vill se frister överst.

Implementationen ligger i gränssnittet, inte i API:t:

- `src/nuxt/app/utils/dueOrder.ts` - `sortByDueFirst()` (stabil sortering)
- `src/nuxt/app/composables/useDueFirstToggle.ts` - toggle-tillståndet, sparas i
  localStorage under `plan:now-due-first`
- `src/nuxt/app/composables/usePriorityMode.ts` - sorterar de synliga målen
  innan trädet byggs
- `src/nuxt/app/pages/now.vue` - knappen
- `src/nuxt/test/unit/dueOrder.test.ts` - enhetstest

## Varför

Ett eget mål i Plan ("Mål med due-date visas först?") sa att deadlines ska synas
först i Gör nu-listan. Första försöket gjorde det i SQL (ORDER BY med due först),
men det blev "jättejobbigt att det alltid kom först" - en fast regel i API:t går
inte att stänga av. En toggle i gränssnittet löser det: API:t behåller sin
ordning, och vyn bestämmer.

## Konsekvens

- Kriterierna för vilka mål som syns är oförändrade (ojorda, startdatum passerat,
  beroenden klara, nu/plats-filter) - bara ordningen ändras.
- Hierarkin är orörd: ett childmål med deadline ligger kvar under sin förälder.
  "Högst upp" gäller inom varje nivå, inte överst i hela trädet.
- Med toggle:en på döljs den manuella ordningen för mål med deadline (de
  sorteras efter datum). Mellan mål med samma deadline gäller den manuella
  ordningen.
- Skrivbordswidgeten läser samma endpoint men har ingen toggle - den visar
  alltså alltid den manuella ordningen.

## Deploy

Ändringen pushades till `main` (commit `cd2e38a`). CI bygger nuxt-imagen och
öppnar PR:t "chore: update plan-app image" - en bot mergar det, Flux synkar och
rollar ut podden. Verifierat i dev (localhost:3100): knappen syns, `aria-pressed`
går från `false` till `true` och första raden byter från det manuellt sorterade
målet till målet med närmaste deadline.

Notera: dev-klustret kör CI-imagen från `main` (Flux sköter deploymenten), så
en lokal Tilt-build syns inte förrän PR:t mergats - Flux skriver tillbaka en
lokal `kubectl set image`.

## Verktyg

- `go build` och `cargo` behöver `GOTMPDIR`/`TMPDIR` peka på en katalog i $HOME -
  /tmp har en kvot som ger "disk quota exceeded".
- `sccache` går sönder med /tmp-kvoten: bygg Rust med `RUSTC_WRAPPER=` (tomt).
- Ny Nuxt-komposable måste följas av `bun run postinstall` (nuxt prepare), annars
  hittar vue-tsc inte den i autoimporterna.
