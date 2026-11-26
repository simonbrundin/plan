---
title: AA Data API v2/data pensionerad - dashboard migrerar till /api/v2/language/models
type: note
tags: [bug, dashboard, artificialanalysis, api, projekt:simon]
created: 2026-11-26
updated: 2026-11-26
status: active
related: [notes/dashboard-ranking-anvander-benchmark-matchade-kostnader]
---

2026-11-25: Artificial Analysis pensionerade /api/v2/data/* (410 Gone sedan 4 nov 2026). Ersattning: /api/v2/language/models (Pro+) och /api/v2/language/models/free (Free, samma svarformat for alla nycklar, paginerat 200/sida, 4 sidor = 766 rader).

Varför detta spelade roll: dashboard har en Free-nyckel (NUXT_ARTIFICIAL_ANALYSIS_API_KEY), och Free-nivån har bara 100 requests per 24h-fönster - en full modelllista kostar 4 requests. Därför cachas listan nu 6h i Nitro 'cache'-mounten (app/.nuxt/cache/aa) + i minnet, så Tilt-omstarter inte brinner kvoten.

Viktiga fynd: (1) nytt wire-format - performance ligger i performance{} och performance_v3{}, pricing saknar price_1m_blended_3_to_1 for Free (approximeras nu som (3*input+output)/4), model_creator saknar slug (sluggas), intelligence index + cost_per_task finns kvar. (2) artificial_analysis_coding_index är alltid null - AA publicerar inte längre någon Coding Index (deprecated), så coding-skor i dashboard kommer från BenchLM/terminal-bench. (3) median-performance nullas ut 2026-12-10, fallback till performance_v3-medel.

Ändringar: src/nuxt/server/utils/artificialAnalysis.ts (ny normalisering + cache + paginering) och src/nuxt/server/api/models/stats.get.ts (använder fetchAAModelCount).

## Relaterade

- [[notes/dashboard-ranking-anvander-benchmark-matchade-kostnader]]

