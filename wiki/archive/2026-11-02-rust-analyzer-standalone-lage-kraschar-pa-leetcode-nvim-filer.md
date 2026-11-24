---
title: rust-analyzer standalone-läge kraschar på leetcode.nvim-filer
type: note
tags: [rust, neovim, rust-analyzer, bug, leetcode, projekt:rust]
---

Problem: När man öppnade en leetcode.nvim-lösning i Neovim sade health-checken
"rust-analyzer (client 2) health status is [error]: cargo check failed to start".

Rotorsak (hittad genom att logga cargo-anropen med en shim):
leetcode.nvim sparar lösningar som FRAGMENT utan Cargo.toml, t.ex.
"1.two-sum.rs" med bara "impl Solution { ... }". rust-analyzer startar då i
standalone/detached-läge och kör
  cargo clippy --manifest-path <fil>.rs -Zscript ...
Flaggan -Z (crate-attr=feature(frontmatter)) kräver en nightly-rustc, och den
rustc som hittas är stable 1.99.1 -> "the option `Z` is only accepted on the
nightly compiler".

Viktigt lärdom om maskinen: mise-shims låser toolchain till den installerade
versionen (RUSTUP_TOOLCHAIN=1.99.1) i alla barnprocesser och ignorerar
RUSTUP_TOOLCHAIN i miljön. Därför gick det inte att "bara köra nightly":
- mise-shim -> stable, även om RUSTUP_TOOLCHAIN=nightly sattes i env
- rustup-proxyn i ~/.cargo/bin respekterar RUSTUP_TOOLCHAIN
- cargo hittar rustc via PATH och ignorerar RUSTC i script-läge
Att peka om cargo.path till proxy + extraEnv RUSTUP_TOOLCHAIN=RUSTC=nightly
räckte alltså inte.

Lösning (vald): stäng av standalone-läget för rust-analyzer:
  { "mrcjkb/rustaceanvim", opts = { server = { standalone = false } } }
Då startar rust-analyzer bara i riktiga cargo-projekt. Filerna i
~/.local/share/nvim/leetcode får ingen LSP-client -> inget health-fel.
Reserven: de går inte att kompilera enskilt ändå (saknar struct Solution och
fn main), och LeetCode-körningen sker server-side.

Verifierat: leetcode-filen -> 0 klienter, inget health-fel; /tmp/hello-rs ->
1 klient (rust-analyzer) som fungerar.
