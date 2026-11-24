---
title: rust-analyzer: tysta health-status-notisen istället för att stänga av standalone
type: note
tags: [rust, neovim, rust-analyzer, leetcode, bug, projekt:rust]
created: 2026-11-02
updated: 2026-11-17
status: active
related: [notes/sa-dampas-bada-rust-analyzer-popupen-no-project-root-found]
---

Fas 2 av samma problem: när standalone-läget stängdes av (server.standalone =
false) fick man i stället popup-felet "No project root found. Not starting
rust-analyzer client because detached/standalone mode is disabled."

Bättre lösning: behåll standalone-läget (default) och dämpa notiserna:
  { "mrcjkb/rustaceanvim", opts = { server = { status_notify_level = false } } }
Notiserna kommer från rust-analyzers "rust-analyzer/serverStatus"-event via
rustaceanvims server_status.lua, och config.server.status_notify_level
(default 'error') styr dem.

För att slippa cargo-script-körningen i leetcode-katalogen finns en giltig
config-nyckel: rust-analyzer.toml med "checkOnSave = false" (OBS: "check.enable"
finns INTE i schemat -> "invalid config value: check/enable: unexpected field").

Fil: ~/.local/share/nvim/leetcode/rust-analyzer.toml
  checkOnSave = false

Verifierat: leetcode-filen -> 1 klient (LSP funkar), 0 notiser; riktigt
cargo-projekt -> 1 klient, fungerar som vanligt.

## Relaterade

- [[notes/sa-dampas-bada-rust-analyzer-popupen-no-project-root-found]]

