---
title: IMAP-sessionen i mailagenten är inte trådsäker - parallella listanrop förstör anslutningen
type: note
tags: [bug, mail, imap, projekt:pappa]
---

Fynd under felsökning av .eml-expansionen: mailagenten håller EN IMAP-anslutning
per konto och delar den mellan alla HTTP-anrop. Två samtidiga listanrop
(interleavade Select/Fetch) korrupterar anslutningen, och felet som kommer tillbaka
är "failed to select folder ...: short write".

"short write" finns inte i isConnectionError-listan i src/go/mcp/session.go, så
sessionen markeras inte som trasig - varje nytt anrop failar med samma 500 tills
podden startas om.

Inte fixat här (utanför uppdragets scope), men värt att veta:
- Testa inte API:t parallellt mot ett konto.
- Att lägga till "short write" i isConnectionError skulle få sessionen att
  återanslutas automatiskt i stället för att stå död.
