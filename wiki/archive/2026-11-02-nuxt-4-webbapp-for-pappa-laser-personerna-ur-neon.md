---
title: Nuxt 4-webbapp för pappa läser personerna ur Neon
type: note
tags: [nuxt, neon, webbapp, ui, projekt:pappa]
---

Webappen ligger i src/nuxt (Nuxt 4.6 + Tailwind v4 via @tailwindcss/vite) och
läser samma Neon-databas som resten av pappa.

VARFÖR SÅ HÄR:
- En enda anslutningssträng: server/utils/db.ts läser NUXT_DATABASE_URL,
  annars DATABASE_URL, annars src/neon/.env.local. Ingen duplicerad hemlighet.
- postgres.js (inte @neondatabase/serverless) räcker: anslutningen går mot
  Neons poolade endpoint med ssl:require och kanalbindning fungerar direkt.
- Server-endpoint /api/people läser person, organization, person_competence
  och person_source och returnerar även facetter (organisationer,
  kompetenser, orter, senioritet) så att klienten slipper extra rundor.
- Typerna ligger i shared/types/people.ts (Nuxt 4-#shared) eftersom server/
  inte kan importeras av klienten.

GDPR I UI:ET: e-post visas bara om contact_published och inte do_not_contact.
Annars visas status i stället för uppgifter.

VERIFIERAT: dev- och produktionsbygge startar, /api/people returnerar 88
personer, sök/filter/sortering fungerar i browsern (playwright), nuxi
typecheck är rent.

LÄRDOM: en .ts-fil i app/composables får INTE innehålla <script setup>-taggar
(nuxi-templaten skapar .vue-filer). Skapa composables som rena .ts-filer.
