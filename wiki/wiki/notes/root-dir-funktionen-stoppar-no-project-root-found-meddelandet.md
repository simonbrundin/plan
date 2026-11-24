---
title: root_dir-funktionen stoppar 'No project root found'-meddelandet
type: note
tags: [rust, neovim, rust-analyzer, leetcode, bug, projekt:rust]
created: 2026-11-02
updated: 2026-11-17
status: active
related: [notes/rust-analyzer-standalone-lage-kraschar-pa-leetcode-nvim-filer]
---

Slutlig fix (fas 4): meddelandet "No project root found" kan undvikas helt vid
källan. rustaceanvims server.root_dir kan vara en funktion
  root_dir = function(file_name, default_get_root_dir)
    return default_get_root_dir(file_name) or vim.fs.dirname(file_name)
  end
Då returnerar cargo.get_config_root_dir() alltid ett root_dir, och grenen
"Not starting..."/"Starting ... detached/standalone mode" i
rustaceanvim/lsp/init.lua körs aldrig -> inget meddelande, varken ERROR eller INFO.

Varför detta behövdes (och varför wrapper-lösningen räckte inte):
LazyVim har en "lazy_notify"-mekanism som byter ut vim.notify under uppstart och
spelar upp de buffrade meddelandena efter ~500 ms. En wrapper som installeras
för sent (t.ex. via FileType-autocmd) missar därför meddelandet, eftersom
uppspelningen går direkt mot den dåvarande vim.notify utanför wrappret.
root_dir-funktionen påverkar inte den vägen -> robust.

Kvarvarande internt fel (men tyst): rust-analyzer kör fortfarande
"cargo check --manifest-path <fil>.rs -Zscript" för leetcode-filerna (cargo
script-läge kräver nightly-rustc). checkOnSave = false i
~/.local/share/nvim/leetcode/rust-analyzer.toml tillämpas men stoppar inte den
första cargo-körningen. Popupen döljs med server.status_notify_level = false.
Att få bort även det inre felet kräver att rust-analyzer kör nightly
(cargo.extraEnv: PATH med ~/.cargo/bin först + RUSTUP_TOOLCHAIN=nightly), vilket
da blir globalt för alla rust-analyzer-projekt.

Verifierat: leetcode-filen -> 1 klient, 0 "No project root"-notiser;
cargo-projekt -> 1 klient med korrekt root (/tmp/hello-rs), 0 notiser.

## Relaterade

- [[notes/rust-analyzer-standalone-lage-kraschar-pa-leetcode-nvim-filer]]

