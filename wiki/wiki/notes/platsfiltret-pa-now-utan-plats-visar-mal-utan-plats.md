---
title: Platsfiltret på /now: "Utan plats" visar mål utan plats
type: note
tags: [feature, now, goals, places, projekt:plan]
created: 2026-11-27
updated: 2026-11-27
status: active
related: [notes/platsvaljaren-flyttade-till-leader-p, notes/skrivbordswidget-som-visar-gor-nu-malet]
---

Symtom: med filtret 'Görbar här' på och 'Ingen plats vald' i rullmenyn 'Jag är här' tömdes listan - isDoableAtPlaceOrUnplaced returnerade false för alla mål när placeId var null.

Ändring (två delar):

1. isDoableAtPlaceOrUnplaced (src/nuxt/app/utils/places.ts) returnerar nu true för mål utan platser även när placeId === null, och false för mål med platser. 'Utan plats' + platsfilter visar alltså de mål som inte är kopplade till någon plats.
2. Platsfiltret är nu PÅSLAGET som standard (useState "now-filter-current-place", default true) och knappen heter 'Filtrera på plats'. Avstängt = visa alla mål oavsett plats.

Varför: ett mål med plats hör till en specifik plats och ska synas när den platsen är vald; ett mål utan plats är inte bundet till någonstans. Den gamla regeln (null -> allt falskt) gav tom lista, vilket uppfattades som "filteret är trasigt". Filter default ON gör appen lika widgeten, som alltid filtrerar på plats.

Resultat-tabell på /now (med filtret på):
- Utan plats -> endast mål utan plats
- Vald plats X -> mål på X + mål utan plats (OR-semantik)
- Filter av -> alla mål

Nu-filtret (now-flaggan) gäller förstås fortfarande, listan är alltså snittet av nu-märket och platsfiltret.

Paritet: skrivbordswidgetens nowlist.js (~/.config/quickshell/plan-goal/nowlist.js) har samma regel - ingen vald plats -> bara mål utan platser. Widgeten hade tidigare filtrerat bort ingenting när ingen plats var vald.

Verifierat: bun test test/unit (265 pass), bun run test (89 scenarios), bun run typecheck (gaten godkänd, inga nya fel), bun run build.

## Relaterade

- [[notes/platsvaljaren-flyttade-till-leader-p]]
- [[notes/skrivbordswidget-som-visar-gor-nu-malet]]
