---
title: MailMaster: /api/v1 gav 'Failed to fetch' - Go-servern panickade och saknade CORS
type: note
tags: [bug, mailmaster, go, cors, projekt:simon]
created: 2026-10-28
updated: 2026-10-28
status: active
---

Symtom: frontend (Nuxt, port 8400) fick '[GET] http://localhost:8401/api/v1/accounts/.../emails?limit=50: <no response> Failed to fetch'.

Rotorsak 1: main.go skapade servern med mcp.NewServer(nil, nil, nil). ensureConnection() dereferencerade s.cfg.Mail och kraschade med nil pointer panic -> Go stängde anslutningen utan svar -> webbläsaren rapporterar 'Failed to fetch'. Serverloggen visade 'http: panic serving ...: runtime error: invalid memory address or nil pointer dereference' i mcp/server.go:97. Introducerad i refaktoreringen d9d66a2 'refactor: consolidate HTTP server into MCP package', som tog bort config-loading ur main.go.

Rotorsak 2: ingen CORS. Nuxt (8400) och Go-API (8401) är olika origin i webbläsaren, så utan Access-Control-Allow-Origin blockeras svaret.

Fix: main.go laddar config/config.yaml + rules.md + AI-provider igen; ensureConnectionFor(accountName) väljer konto efter namn ur URL:en och returnerar fel (inte panic) om config saknas; withCORS-middleware i mcp/http_server.go (echo Origin, OPTIONS -> 204); frontend encodar konto/mapp-namn med encodeURIComponent.

Fakta: INBOX.Simon är en tom mapp (0 mail) på imap.one.com; INBOX har 2183 mail - tom lista är korrekt svar, inte fel.

Verifierat: curl mot 8401 med Origin-header ger 200 + CORS-headrar; OPTIONS ger 204.
