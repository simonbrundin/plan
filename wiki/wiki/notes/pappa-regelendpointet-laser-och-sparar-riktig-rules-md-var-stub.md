---
title: pappa: regelendpointet läser och sparar riktig rules.md (var stub)
type: note
tags: [bug, mail, pappa, projekt:pappa]
created: 2026-11-16
updated: 2026-11-16
status: active
---

Symtom: /mail/rules i pappa visade alltid stuben '# MailMaster Rules / Define email processing rules here.' Root cause: GET /api/v1/rules i src/go/mcp/http_server.go var hardkodad, och PUT kvitterade utan att spara. Fix: ny src/go/mcp/rules.go - server läser regelfilen (-rules, /app/rules.md i klustret), sparar via atomisk tempfil+rename, validerar innehållet (tomt eller markdown utan regelrubriker -> 400) och byter ut regeluppsättningen i minnet så AI:n använder nya regler utan omstart. Begränsning kvar: filen ligger i containerns filsystem och går förlorad om poden skapas om. Deploy: tilt trigger mailagent (tilt-filen watcher triggar inte på src/go-ändringar).
