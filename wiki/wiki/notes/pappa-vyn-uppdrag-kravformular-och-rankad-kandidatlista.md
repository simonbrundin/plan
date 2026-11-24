---
title: pappa: vyn Uppdrag - kravformulär och rankad kandidatlista
type: note
tags: [pappa, ui, matchning, uppdrag, projekt:pappa]
created: 2026-11-03
updated: 2026-11-03
status: active
related: [notes/pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning, notes/pappa-uppdragsmatchning-delpoang-forklarande-ranking]
---

Efter att ha byggt matchningen i SQL (fn_match_candidates + job_profile) byggdes
även UI:t: sidan /uppdrag med ett kravformulär och en rankad kandidatlista.

Varför i den ordningen: vikterna provades först i SQL mot riktiga data. Då
visade det sig att must-have normalt måste vara EN tagg (tre taggar gav ingen
godkänd kandidat) och att poängen ensam diskriminerar dåligt (många har rätt
ort och senioritet över krav). Därför sorterar listan först på kravtäckning
och sedan på poäng, och därför finns min_must_have ("minst N av M").

Beslut i UI:t:
- Skrivning mot job_profile kräver NUXT_ADMIN_TOKEN, samma token som
  dubblettsammanslagningen. En kravprofil styr vilka kandidater som rankas,
  så en öppen write-endpoint skulle kunna styra listan. Läsning och rankning
  är öppet - det är samma uppgifter som kandidatlistan visar.
- Delpoängen visas per del när kandidaten expanderas. En siffra utan
  förklaring går inte att ifrågasätta.
- Dokumenterad och rollnivåbaserad erfarenhet markeras olika i listan, så att
  en gissad erfarenhet inte läses som en uppgift.
- Taggar väljs ur GET /api/competences, så en must-have är alltid en riktig
  tagg. En tagg som inte finns i ordboken matchar ingenting - det syns inte
  i UI:t annat än som en tom täckning.
- Underkända (täcker inte must-have) hamnar under de godkända, inte borta.
  En människa ska kunna se och överskrida.

Admin-token är nu delad mellan dubblettpanelen och uppdragsvyn via
useAdminToken (sessionStorage, aldrig localStorage).

Kvar: kundkonfliktkontroll, spoken_languages och available_from saknas
fortfarande helt, så språk- och tillgänglighetsdelen är neutral (50).

## Relaterade

- [[notes/pappa-dubblettknapp-mansklig-granskning-och-bestandig-sammanslagning]]
- [[notes/pappa-uppdragsmatchning-delpoang-forklarande-ranking]]
