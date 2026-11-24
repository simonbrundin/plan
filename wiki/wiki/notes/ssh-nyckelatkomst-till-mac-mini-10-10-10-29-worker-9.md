---
title: SSH-nyckelåtkomst till Mac mini 10.10.10.29 (worker-9)
type: note
tags: [ssh, macmini, worker-9, projekt:simon]
created: 2026-11-24
updated: 2026-11-24
status: active
---

Mac mini i datacentret (10.10.10.29, UniFi 'MinisomhorSIMON', MAC a0:ce:c8:d9:0d:02) är samma maskin som worker-9 i Talos-klustret, men bootar just nu macOS 15.7.2 (Talos-partitionen väljs manuellt i bootmenyn). Maskinen kan alltså antingen vara macOS-värd eller Talos-worker - inte båda samtidigt.

MacOS-kontot heter 'simon' (INTE 'simonbrundin' som står i 1Password-itemet 'Mac Mini' - det är bara en etikett). Lösenordet finns i 1Password (op://Private/Mac Mini/password) och användes för att installera nyckeln via ssh-copy-id.

Klientsidan: ~/.ssh/macmini_ed25519 (ed25519) + Host macmini i ~/.ssh/config (HostName 10.10.10.29, User simon, IdentitiesOnly yes). Verifierat: 'ssh macmini' loggar in med nyckel utan lösenord.

FALLGROP: sudo systemsetup -setremotelogin on gav 'invalid command' på den maskinen - funkar istället via GUI (System Settings > General > Sharing > Remote Login) eller 'sudo launchctl enable system/sshd' + 'sudo launchctl load -w /System/Library/LaunchDaemons/ssh.plist'.
