---
title: Så dämpas båda rust-analyzer-popupen (No project root found)
type: note
tags: [rust, neovim, rust-analyzer, leetcode, bug, projekt:rust]
---

Fas 3: rustaceanvim har TVÅ notiser för filer utan Cargo.toml, beroende på
standalone-flaggan:
  standalone = false -> ERROR "Not starting rust-analyzer client because
                             detached/standalone mode is disabled."
  standalone = true  -> INFO  "Starting rust-analyzer client in
                             detached/standalone mode (with reduced
                             functionality)."
Plus en tredje: rust-analyzers serverStatus -> "health status is [error]".

Slutgiltig lösning (alla tre tystade, LSP funkar):
1. standalone lämnas på default (true) -> LSP startar i detached-läge.
2. rustaceanvim opts: server.status_notify_level = false -> tystar health-status.
3. ~/.local/share/nvim/leetcode/rust-analyzer.toml: "checkOnSave = false"
   -> cargo-script-körningen (som kräver nightly) sker inte.
4. "No project root found"-notisen dämpas genom att wrappa vim.notify i en
   FileType-autocmd (pattern "rust"). Viktigt: autocmden måste registreras
   under lazy.setup() (dvs i plugins/*.lua toppnivå-kod) så att den körs
   FÖRE rustaceanvims LSP-start, annars hinner notisen iväg.
   Wrapper fångar bara meddelanden som matchar "No project root found".

Verifierat via snacks.notifier.get_history(): ingen "No project root found"
i historiken, vanliga notiser fungerar, 1 rust-analyzer-klient i
leetcode-filen och i riktiga cargo-projekt.
