---
title: Ctrl-Backspace i foot skickar Ctrl-W till Pi
type: note
tags: [bug, terminal, pi, projekt:dotfiles]
---

Foot skickade Ctrl+Backspace som rå byte 0x08. Pi på Linux tolkar inte 0x08 som ctrl+backspace, och tmux 3.8 vidarebefordrar inte Kitty-keyboard-protokollet. Foot 1.28.0 verifierar därför text-bindningen \x17=Control+BackSpace i [text-bindings], vilket skickar Ctrl+W (0x17). foot -C validerar konfigurationen. Normal Backspace lämnas oförändrad.
