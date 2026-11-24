---
title: Ikonväljaren i plan: /api-icons-routen var borttagen och ikonprovider none gjorde att inga ikoner renderades
type: note
tags: [bug, nuxt, icon, plan, projekt:plan]
created: 2026-11-16
updated: 2026-11-17
status: active
related: [notes/plan-cachad-spa-skal-gjorde-att-webblasaren-korde-gammalt-js-bundle-med-ikon-provider-none]
---

Symtom: i Nuxt-applet gick det inte att välja ikon för ett mål - IconPicker.vue visade 'Ingen ikon matchade "muscle"' oavsett sökning.

Rotorsak 1: commit ea47707 ('refactor: remove Nuxt server routes, use Go API instead') raderade server/api/icons.ts (Iconify-sökning) ur Nuxt, men IconPicker.vue fortsatte att anropa /api/icons. Sökningen träffade ingen route - Nitro svarade med SPA-html, JSON-parsningen kastade och listan blev tom.

Rotorsak 2: nuxt.config.ts hade icon.provider = 'none'. @nuxt/icon laddar då ingen data alls (client-bundlen blev 'export function init() {}'), så även dynamiska ikoner (goal.icon) renderades inte. Detta var dolt eftersom komponenten bara visar text.

Fix: (1) src/nuxt/server/api/icons.get.ts återupptog Iconify-sökningen (q-param, limit 64, fel -> 502 så att klienten kan skilja fel från tom träff); (2) icon.provider = 'iconify' - klienten hämtar ikonerna från api.iconify.design, vilket också gör att valfri ikon från IconPicker kan renderas; (3) IconPicker.vue skiljer nu på 'sökningen misslyckades' och 'ingen träffar'.

Varför: ikonsökning och ikonrendering är UI-behov utan affärslogik - de ska inte flyttas till Go API:t när Nuxt-routes tas bort. Framtida refaktoreringar som raderar Nuxt-servercode måste kontrollera att klienten inte anropar en route som försvinner.

Verifierat: /api/icons?q=muscle returnerar iconoven:muscle etc., ikoner renderas i headless chromium (CSS-läge, --svg-data-URI satt), cucumber-testerna (6 scenarios) passerar.

## Relaterade

- [[notes/plan-cachad-spa-skal-gjorde-att-webblasaren-korde-gammalt-js-bundle-med-ikon-provider-none]]

