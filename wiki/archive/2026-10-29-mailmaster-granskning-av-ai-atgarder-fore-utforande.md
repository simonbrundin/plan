---
title: MailMaster: granskning av AI-åtgärder före utförande
type: note
tags: [feature, mailmaster, frontend, backend, projekt:mailmaster]
---

Processa mejl-knappen kör nu inte längre åtgärder direkt. Nytt tvåstegsflöde i UI:t:

1. POST /api/v1/accounts/{name}/analyze - AI:t analyserar mejlen utan att utföra något och svarar med vilken regel som matchade (REGEL-fältet i AI-svaret), motivering, konfidens och en lista med rekommenderade åtgärder. Resultatet sparas serverside (Server.analyses, nyckel konto|mapp|uid).
2. POST /api/v1/accounts/{name}/decision - approve/deny. Vid approve körs exakt de föreslagna åtgärderna med de argument AI:t föreslog (sparade serverside, så klienten kan inte ändra vad som händer).

Varför: inget ska ske i brevlådan utan att användaren godkänner det. AI-fel eller en saknad REGEL-rad ger istället en förklaring i listan.

Frontend: EmailList.vue visar nu per mejl en beskrivning av regelmatchningen, en ruta med rekommenderade åtgärder samt OK/Neka-knappar till höger. ResultsModal.vue togs bort.

Viktigt: actions.Executor.ExecuteActions kör bara de angivna åtgärderna (Execute letar upp extra åtgärder i motiveringstexten).
