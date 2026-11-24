---
title: Feedback-loop: AI uppdaterar mailreglerna och analyserar om mejlet
type: note
tags: [mailagent, ai, feedback, pappa, session, projekt:pappa]
---

2026-11-19 (pappa/mailagent): Istället för en "Neka"-knapp kan användaren nu starta en konversation med AI:t om vad som inte fungerade med en analys eller de rekommenderade åtgärderna.

Varför: En "Neka"-knapp kastar bort informationen - AI:t får veta att något var fel men inte vad. Med en konversation kan reglerna justeras så att samma fel inte upprepas, och det är just feedback-loopen som saknades i specs/ai-mail-processor.md ("Framtida utökningar").

Hur det fungerar:
- Ny endpoint POST /api/v1/accounts/{konto}/feedback (src/go/mcp/feedback.go).
- AI:t får systeminstruktion + nuvarande regler + analysen + mejlets innehåll + användarens fråga.
- Svaret förväntas innehålla en kort förklaring och de FULLSTÄNDIGA nya reglerna i ett ```markdown-block.
- Blocket valideras med rules.ParseMarkdown innan det sparas - en regeluppsättning som inte kan tolkas tas inte över, de fungerande reglerna ligger kvar.
- Sparas det, analyseras mejlet om med de nya reglerna och den nya rekommendationen skickas tillbaka till UI:t (samma analys-objekt som godkännande sedan använder).
- Saknas systeminstruktionen i AI:ts förslag behålls den gamla - annars förlorar analysen sina instruktioner.
- Dialogen sparas per mejl (konto|mapp|UID) så flera frågor bygger på varandra. Ny "Processa mejl"-körning rensar konversationerna.

Implementationen:
- ai.OpenAIProvider.Chat (src/go/ai/chat.go) - samma modell (gpt-5.6-luna) som analysen, men fri text i stället för det strikta ÅTGÄRD-formatet. responsesRequest.Input ändrades till interface{} så en hel konversation kan skickas.
- UI: FeedbackDialog.vue (modal) + knappen "Justera med AI" i EmailReview.vue och EmailList.vue. Neka-knappen ersatt i båda vyerna; /decision-endpointen finns kvar för agenter.

Verifierat end-to-end mot riktig IMAP + riktig AI (lokal binär, regelfil i /tmp så inte användarens regler rördes): AI:t lade till regeln "Mejl från info@upgraded.se", skärpte "Sundsvall eller Remote", och mejlet analyserades om och matchade den nya regeln.

Fynd: vue-tsc/typecheck fångar INTE en trasig v-if/v-else-kedja i en Vue-template - bara `nuxt build` gör det. FeedbackDialogen fick först placeras mellan v-else-if-blocken och bröt bygget.
