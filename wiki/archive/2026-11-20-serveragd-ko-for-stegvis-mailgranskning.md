---
title: Serverägd kö för stegvis mailgranskning
type: note
tags: [mail, go, granskning, kö, projekt:pappa]
---

2026-11-19 (uppdaterad): Efter att "Processa mejl" körts i pappa/appen visas ett mejl i taget i en granskningsvy (src/nuxt/app/components/mail/EmailReview.vue) i stället för hela brevlådelistan.

Varför: listan med upp till 50 mejl gjorde det svårt att fokusera på det AI:n faktiskt föreslår åtgärder för. Granskningen visar i ordning: mejlets titel, resonemanget (AI:s motivation), rekommenderade åtgärder, brevtexten, och till höger Acceptera/Neka.

Viktiga implementationdetaljer:
- Servern äger kön. POST /api/v1/accounts/{konto}/analyze utan "uid" bygger en kö av brevlådans UID:n och analyserar BARA det första mejlet. Med "uid" analyseras ett specifikt mejl i kön - och bara om UID:t finns i kön, annars avvisas anropet. Det gör att ett mejl aldrig kan godkännas utan att ha granskats.
- Varje analys sparas separat i servern (storeAnalysis per konto+mapp) så att tidigare analyser överlever när nästa mejl analyseras. En ny körning rensar gamla analyser (clearAnalyses) och regelkonversationer.
- Frontendens useMailEmailProcessing håller queue/activeIndex och cachear analyser per UID så att "Föregående" inte kostar nya AI-anrop.
- Brevtexten hämtas lazy via GET /api/v1/accounts/{konto}/folders/{mapp}/emails/{uid}, inte via list-endpointen - den innehåller bara body för expanderade .eml-mejl.
- Efter ett beslut (Acceptera/Neka) analyseras nästa mejl automatiskt. "Nästa" hoppar över ett mejl manuellt (det markeras som hoppat och kommer inte tillbaka i kön), "Föregående" går tillbaka till ett redan granskat mejl.
- Lista: "Mejl X av Y · N kvar" där N är antal mejl som fortfarande saknar beslut. När alla är genomgångna visas ett "Klart"-läge.

Verifierat med Playwright mot dev-klustret: 10 mejl i kön krävde exakt 10 analyze-anrop (ett per mejl, aldrig alla på en gång), Acceptera/Neka flyttar kön framåt, och klart-läget nås efter sista beslutet.
