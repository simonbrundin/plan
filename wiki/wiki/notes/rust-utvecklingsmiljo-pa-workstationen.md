---
title: Rust-utvecklingsmiljö på workstationen
type: note
tags: [rust, toolchain, mise, neovim, projekt:rust]
created: 2026-11-02
updated: 2026-11-17
status: active
related: [notes/root-dir-funktionen-stoppar-no-project-root-found-meddelandet, notes/rust-learning-repot-lades-in-som-submodule-mot-egen-fork]
---

Vi satte upp en komplett Rust-miljö på maskinen (2026-11-02).

Utgångsläge: rustc/cargo/rustup 1.99.1 via mise-shims, clippy+rustfmt, men saknade
rust-src, rust-analyzer, wasm-target, nightly och alla moderna cargo-verktyg.

Viktigaste fyndet: rust-analyzer-komponenten saknades i toolchain. Mise-shims
pekar på rustup-proxyn, som utan komponenten faller tillbaka på shim -> "infinite
recursion detected" i nvim-lsp.log, och LSP startar aldrig. Lösning:
rustup component add rust-analyzer.

Ändringar:
- rustup: rust-src, rust-analyzer, wasm32-unknown-unknown, nightly (med miri)
- ~/.cargo/config.toml: sccache som rustc-wrapper, clang+lld som linker
  (mold kunde inte installeras - saknar sudo)
- cargo-verktyg via cargo-binstall: bacon, cargo-nextest, cargo-llvm-cov,
  cargo-audit, cargo-deny, cargo-watch, just, sccache, cargo-edit, cargo-expand,
  cargo-bloat
- Neovim (dotfiles-repo ~/repos/dotfiles): ny lua/plugins/rust.lua som aktiverar
  LazyVims lang.rust-extra (rustaceanvim, crates.nvim, treesitter rust/ron,
  codelldb) samt conform-formatters for rust (rustfmt) och toml (taplo)

Lärdom: `curl https://sh.rustup.rs | sh` interaktivt failar pa maskinen eftersom
mise-shims ligger i PATH - installern anser att Rust redan finns och avbryter med
"cannot install while Rust is installed". Anvand `-y --no-modify-path`, da
uppdateras den befintliga toolchain utan problem.

## Relaterade

- [[notes/rust-learning-repot-lades-in-som-submodule-mot-egen-fork]]
- [[notes/root-dir-funktionen-stoppar-no-project-root-found-meddelandet]]
