---
title: Mailgranskningen visar ett mejl i taget
type: note
tags: [mail, nuxt, granskning, ui, projekt:pappa]
created: 2026-11-19
updated: 2026-11-19
status: active
---

2026-11-19: Efter att 'Processa mejl' körts i pappa/appen visas numera ett mejl i taget i en granskningsvy (src/nuxt/app/components/mail/EmailReview.vue) i stället för hela brevlådelistan.

Varför: listan med upp till 50 mejl gjorde det svårt att fokusera på det AI:n faktiskt föreslår åtgärder för. Granskningen visar i ordning: mejlets titel, resonemanget (AI:s motivation), rekommenderade åtgärder, brevtexten, och till höger Acceptera/Neka.

Viktiga implementationdetaljer:
- Kön byggs ur analyssvaret i useMailEmailProcessing (queue) - mejl med ett beslut (approved/denied) lämnar kön, så nästa mejl visas direkt efter ett beslut.
- Brevtexten hämtas lazy via GET /api/v1/accounts/{konto}/folders/{mapp}/emails/{uid} (fanns redan i Go-API:t), inte via list-endpointen - den innehåller bara body för expanderade .eml-mejl.
- Ingen ändring behövdes i Go-tjänsten; listvyn (EmailList) ligger kvar och syns när man inte är i granskningsläge.

Verifierat med Playwright mot dev-klustret: vyn renderar korrekt, navigering Föregående/Nästa fungerar och Neka flyttar kön 10 -> 9.
