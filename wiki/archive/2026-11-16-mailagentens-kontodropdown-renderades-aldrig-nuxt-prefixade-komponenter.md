---
title: Mailagentens kontodropdown renderades aldrig - Nuxt-prefixade komponenter
type: note
tags: [bug, nuxt, mail, projekt:pappa]
---

2026-11-16: pappa:s mailagent (flyttad från mailmaster) hade en kontodropdown i koden, men den syntes ALDRIG i webbappen.

## Orsak

Flytten in i pappa (bd271e7) flyttade komponenterna från src/nuxt/app/components/ till src/nuxt/app/components/mail/. Nuxt registrerar komponenter i underkataloger med katalognamnet som prefix - AccountSelector.vue i components/mail/ blir MailAccountSelector. Men sidan (app/pages/mail/index.vue) refererar <AccountSelector />, <EmailControls /> och <EmailList /> utan prefix.

Vue hittade dem aldrig. Tre tecken på <!----> i SSR-HTML:t, och noll account-bar i det byggda outputet - dropdownen, mapparen, knapparna och maillistan var samtliga osynliga. Ingen kompileringsvarning, ingen test som fångade det.

## Fix

components: [{ path: '~/components', pathPrefix: false }] i nuxt.config.ts.

## Fällan att se upp för

Nuxt-prefixet är SILENT. Ett komponentnamn som inte matchar ger ingen
byggvarning - bara en tom platshållare. Efter en flytt av komponenter till
en underkatalog ska man alltid kolla .nuxt/components.d.ts (eller köra
) och jämföra
namnen mot de referenser sidorna gör.

Verifiering: SSR-HTML:t ska innehålla account-bar, folder-select och
Hämta
