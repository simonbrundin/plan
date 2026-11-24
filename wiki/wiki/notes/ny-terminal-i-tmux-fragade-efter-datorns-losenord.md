---
title: Ny terminal i tmux frågade efter datorns lösenord
type: note
tags: [bug, shell, 1password, op, tmux, projekt:simon]
created: 2026-11-20
updated: 2026-11-20
status: active
---

Varje ny shell (tmux-fönster/tab, ny terminal) startade ~/.config/secrets/load.sh, som KÖRDE `op read op://Private/Neon API key/credential` varje gång. 1Password CLI:n har delad låsstatus med desktop-appen (settings.json: developers.cliSharedLockState.enabled=true). När appen är låst triggar `op read` appens 'system auth'-upplåsning via polkit -> PAM -> quickshell (omarchy-shell) visar en autentiseringsdialog som ber om DATORNS lösenord. Därför frågades lösenordet varje gång en ny terminal/tab öppnades. Anropet tog dessutom 8-60s (native messaging-anslutning till appen), vilket gjorde att terminalen verkade hänga.

Fix (2026-11-20): load.sh och nushell secrets.nu läser nu CACHEN (~/.config/secrets/neon_api_key, mode 600) först och anropar op BARA om cachen saknas. Cachen är utanför git. Manuell uppdatering via 'secrets refresh' (ny funktion i bash, befintlig action i nushell).

FÖRBEHÅLL: tokenet åldras inte automatiskt. Efter en nyckelrotation måste 'secrets refresh' köras manuellt, annars använder MCP-klienterna (pi, opencode, goose, gemini) den gamla NEON_API_KEY.

Liknande fynd under vägen: omarchy-sleep-lock.service startar om omedelbart hundratals gånger (Restart=always) eftersom /usr/bin/omarchy-system-sleep-monitor inte kan ansluta till systemets D-Bus. Separat problem, inte relaterat till lösenordsprompten.
