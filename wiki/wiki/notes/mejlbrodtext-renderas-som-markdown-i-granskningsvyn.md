---
title: Mejlbrödtext renderas som markdown i granskningsvyn
type: note
tags: [mail, markdown, nuxt, säkerhet, projekt:pappa]
created: 2026-11-19
updated: 2026-11-19
status: active
---

2026-11-19: I granskningsvyn (EmailReview.vue) visas mejlets brödtext nu som renderad markdown i stället för rå text i ett pre-block. Varför: mejlens BODY[TEXT] innehåller ofta markdown (rubriker, listor, länkar) som såg trasigt ut. Lösning: ny komposabel src/nuxt/app/composables/useMarkdown.ts bygger på markdown-it (nytt beroende ^15) med html:false, breaks:true, linkify:true; länkar får target=_blank och rel=noopener noreferrer nofollow; stylen ligger i scoped CSS med deep() eftersom innehållet sätts in med v-html. Säkerhet: html:false escapar all rå HTML i brevet och markdown-it blockerar javascript:-länkar. Verifierat med Playwright mot dev-klustret: rubriker, listor (10 li) och länkar (18 st) renderas korrekt i riktiga mejl.
