---
title: Pi-extension: minimax-quota-guard väntar ut 5h-fönstret och visar pace
type: note
tags: [pi, extension, minimax, kvota, projekt:dotfiles]
---

Ny extension ~/.pi/agent/extensions/minimax-quota-guard.ts (dotfiles-repo, pi/.pi/agent/extensions/), plus tester i pi/tests/.

Varför: MiniMax Token Plan har ett rullande 5-timmarsfönster och en veckofönster. När 5h-fönstret tar slut (HTTP 429, felkod 2056 'usage limit exceeded') dog körningen. Nu spärrar extensionen innan körningen (before_agent_start) och återupptar automatiskt efter nollställning; fel som uppstår mitt i en körning fångas via message_end och körningen återupptas efter fönstret öppnat. Veckokvoten slut hanteras inte (nytt 5h-fönster hjälper inte).

Viktigt: felmeddelandet skrivs om till 'kvota är slut' eftersom @uzunkonak/pi_retry annars matchar samma feltext och båda extensionerna skulle återuppta jobbet (dubbelkörning). Kod 2062 (tillfällig RPM-limit) och 2067 (ingen nollställningstid) räknas inte som 5h-fönster.

Pace-indikator (widget belowEditor, vänster): andel av kvoten som använts jämfört med andel av 5h-fönstret som gått. Under 95 % av tak = grön, 95-105 % = gul, över 105 % = röd.

Data: GET /v1/token_plan/remains (www.minimax.io), nyckel via MINIMAX_API_KEY. Procenttalen är kvarvarande %, tidsstämplar epoch-ms, felstatus i base_resp.status_code trots HTTP 200. Befintlig @cwillam/pi-minimax-usage pollar samma endpoint för footer-status.
