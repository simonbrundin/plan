---
title: Pi: tog bort pi-rate-limit och pi-pattern-retry, fixade omarchy temasy
type: note
summary: Ändring 2025-xx-xx i dotfiles-repo, pi-agentens extension-uppsättning:
tags: [pi, extensions, omarchy, projekt:simon]
created: 2026-10-23
updated: 2026-10-23
status: active
---

Ändring 2025-xx-xx i dotfiles-repo, pi-agentens extension-uppsättning:

- Tog bort paketen pi-rate-limit (lokal fork i vendor/ + patch-scriptet
  scripts/patch-npm-extensions.mjs) och pi-pattern-retry. Båda var överlappande
  retry-motorer tillsammans med @uzunkonak/pi_retry, som är den som behölls.
  Skälet: pi_retry har live-countdown och förstår prenumerationsfönster
  ("resets 11:30pm"), de andra två gav dubbla timers.
- Uppdaterade extensions/omarchy-system-theme.ts: den läste
  ~/.config/omarchy/current/theme/light.mode som inte finns på maskinen, så
  den var död kod. Nu läser den mode = "light"/"dark" från
  ~/.local/state/omarchy/current/theme/colors.toml (där Omarchy materialiserar
  det aktiva temat), med temats namn som fallback.

Lärdom: Omarchy materialiserar aktuellt tema i ~/.local/state/omarchy/current/,
inte i ~/.config/omarchy/current/ (som inte existerar).

## Relaterade
