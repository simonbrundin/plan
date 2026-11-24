---
title: Sessionen saknade executor - varje godkänd åtgärd kraschade med 502
type: note
tags: [bug, mail, mcp, projekt:pappa]
---

Symtom: i /mail gav "ok" på en föreslagen åtgärd 502 Bad Gateway från
Nuxt-proxyn. Serverloggen visade "http: panic serving ...: runtime error:
invalid memory address or nil pointer dereference" med stacken
actions.(*Executor).archive -> mcp.(*Server).httpDecideEmail.

Rotorsak: accountSession skapades utan executor. NewExecutor anropades
hela repot inte (bara deklarationen i actions/executor.go fanns), så
session.executor var nil. Varje anrop till en metod på ett nil *Executor
går inte tillbaka i Go - metoden körs med nil-receiver, och första
fältåtkomsten (e.imapClient) ger nollpekare. Paniken dog i
http-serven, som stängde anslutningen - nginx-proxyn såg det som 502.

Fix:
- ensureConnectionFor bygger sessionen via newSession, som ger varje
  session sin egen executor (actions.NewExecutor med sessionens
  imap-klient och fallback-konfig) plus SMTP-credentials.
- withRecovery-middleware runt HTTP-handlern: en panik ger nu 500 med
  JSON-body i stället för en avbruten anslutning.
- Nollpekarskydd i actions (nil email, nil cfg, nil imapClient).
- Regressionstest: mcp/session_test.go - en session MÅSTE ha en executor.

Lardom: en struct-fält som aldrig sätts ger inget kompileringsfel och
ingen test röjer felet förrän exekveringsvägen nås. Sök efter
fältanvändning (inte bara deklarationen) när ett fält verkar dött.
