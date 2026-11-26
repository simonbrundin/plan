---
title: Dokumentationssidor i gränssnittet (/docs)
type: note
tags: [dokumentation, nuxt, projekt:plan]
created: 2026-11-26
updated: 2026-11-26
status: active
---

Plan har fått en inbyggd dokumentation: undersidor på /docs/[undersida] + översikt på /docs.

Varför: AGENTS.md, docs/ och repot som helhet var spretigt - det fanns ingen samlad, läsbar beskrivning av hur appen fungerar eller hur agentsystemet fungerar.

Hur: markdown-filer i src/nuxt/app/content/docs/*.md, importeras via import.meta.glob (?raw) i app/composables/useDocs.ts, renderas med marked i app/utils/docs.ts (rubriker får id:n för innehållsförteckning). Sidor: app/pages/docs/index.vue och app/pages/docs/[slug].vue. Typografin ligger i app/assets/css/docs.css eftersom innehållet renderas via v-html.

Innehåll: arkitektur, gränssnitt och tangenter, mål och datamodell, agentsystemet (faser, profiler, gates, human gate, kostnad), agentkörningar och godkännanden (Kubernetes Job, säkerhetsgränser), API och autentisering, utveckling och drift.

Fälla: @nuxt/content finns INTE i projektet - markdownen importeras manuellt. marked lades till som direkt dependency (var bara transitiv via semantic-release).
