---
title: Chattmeddelanden från agenten renderas nu som markdown i Plan
type: note
tags: [feature, agent, chat, markdown, projekt:plan]
---

Agentens svar i chattpanelen (src/nuxt/app/components/agent/ChatPanel.vue) visades tidigare som rå text - asterisker, #-rubriker och kodblock syntes som nonsens. Nu renderas assistant-meddelanden som markdown via v-html.

Varför: agenten svarar i markdown, och utan rendering blir svaret svårläst - särskilt listor och kodblock.

Hur: ny app/utils/markdown.ts (marked 15, redan beroende) med renderMarkdown(). Säkerhet: rå HTML escapas (renderer.html), länkar filtreras (isSafeHref tillåter bara http/https/mailto/tel och interna / #) och externa länkar får target=_blank rel=noopener noreferrer. Renderingar cachas (Map, max 200) eftersom chatten ritar om sig under streaming. Typografin ligger i app/assets/css/chat.css (.chat-markdown) och läggs in i nuxt.config.ts css-listan - samma mönster som docs.css, eftersom v-html-innehåll inte kan stylingas med scoped CSS.

Verifierat: 13 unit-tester i src/nuxt/test/unit/markdown.test.ts (bun test), typecheck-gaten utan nya fel, bun run build OK, samt computed-style-check i Chromium mot den byggda CSS-filen (rubriker, listor, inline-kod, kodblock, länkar, citat).
