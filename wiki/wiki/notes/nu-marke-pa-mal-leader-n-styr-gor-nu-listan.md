---
title: Nu-märke på mål (leader+N) styr Gör nu-listan
type: note
tags: [feature, nuxt, go, db, projekt:plan]
created: 2026-11-22
updated: 2026-11-22
status: active
---

Mål har nu en 'now'-flagga (BOOLEAN, samma mönster som 'agent'). Sätts med leader+N på målsidan, /goals och /now, eller via det gröna blixtmärket (NowBadge.vue).

Varför: Gör nu-sidan (/now) visade tidigare mål som bara var påbörjade (started != null). Nu styrs listan av nu-flaggan i stället - filtret 'Nu' är default och 'Alla/Påbörjade/Ej påbörjade' finns kvar som alternativ.

Nyckelbeslut: 'now' är ett eget DB-fält, inte en derivat av 'started'. Det ger friheten att markera något som 'ska göras nu' utan att påbörja det.

Schema: environments/base/db/schema.yaml (deklarativ källa, Atlas Operator applicerar). Kolumnen fanns redan i dev-Neon (default false, inga rader satt).
