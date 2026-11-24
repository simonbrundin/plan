---
title: Neon API-token hamnade i dotfiles-git-historiken
type: note
tags: [säkerhet, dotfiles, secret, projekt:simon]
created: 2026-11-17
updated: 2026-11-17
status: active
related: [notes/losning-1password-backed-env-variabel-for-mcp-tokens]
---

I pi/.pi/agent/mcp.json och opencode/.config/opencode/opencode.json ligger Neon MCP-tokens i klartext, och de committades 2026-11-17 i dotfiles-repot (git-push). repot har redan haft en GitHub PAT i .mac/mcphub/servers.json, så mönstret är inte nytt - men token i historiken bör roteras och flyttas till 1Password/env-referens. Ingen hemlighet skrivs i den här noten.

## Relaterade

- [[notes/losning-1password-backed-env-variabel-for-mcp-tokens]]

