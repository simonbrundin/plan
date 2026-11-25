---
title: Foot-bindningen kräver omstart av foot
type: note
tags: [bug, terminal, pi, projekt:dotfiles]
created: 2026-11-25
updated: 2026-11-25
status: active
---

Konfigurationen \x17=Control+BackSpace validerar korrekt, men foot läser foot.ini vid uppstart. Den aktiva foot-processen startade 10:53, medan foot.ini ändrades 14:49, så den kör fortfarande den gamla keymapen. Slutsats: parse-validering räcker inte som end-to-end-verifiering; öppna ett nytt foot-fönster eller starta om foot och testa därefter.
