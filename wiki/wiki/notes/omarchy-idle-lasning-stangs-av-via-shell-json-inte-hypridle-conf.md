---
title: Omarchy: idle-låsning stängs av via shell.json (inte hypridle.conf)
type: note
tags: [omarchy, hyprland, bug, session, projekt:dotfiles]
created: 2026-11-23
updated: 2026-11-23
status: active
---

Anledning: hypridle.conf i ~/.config/hypr/ hade idle-listener utkommenterade (så skärmen skulle bara låsa vid lock-close), men datorn låste ändå efter ~5 min inaktivitet. Orsak: nyare Omarchy har flyttat idle-logiken till Omarchy-shell (Quickshell) - pluginen shell/plugins/services/idle/Service.qml läser idle.lock/idle.screensaver ur ~/.config/omarchy/shell.json och har en egen IdleMonitor. hypridle.conf-listener påverkar inte den vägen.

Fix: satt idle.lock = 0 och idle.screensaver = 0 i ~/.config/omarchy/shell.json. Då blir idleTimersEnabled falskt, IdleMonitor stängs av och ingen idle-cykel startar - datorn låser inte alls automatiskt. shell.json hot-reloadas (omarchy-shell idle status visar lock=0, lastEvent idle-cycle-cancel: idle-timers-disabled).

Viktigt: låsning vid lock-close (omarchy-system-lid-close, hypridle before_sleep_cmd, omarchy-system-sleep-lock) är separat och fortfarande aktiv - den styrs inte av shell.json.
