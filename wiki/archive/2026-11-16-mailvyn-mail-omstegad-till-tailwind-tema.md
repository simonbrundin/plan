---
title: Mailvyn (/mail) omstegad till Tailwind-tema
type: note
tags: [ui, mail, tailwind, projekt:pappa]
---

Sidorna /mail och komponenterna AccountSelector, EmailControls, EmailList använde handskrivet scoped-CSS med en egen reset (:deep(*)) och en lila gradient, skilt från appens ink/brand-tema i main.css. De är nu omstegade till Tailwind-utilities (card/chip, ink-*, brand-*) och delar header/stil med / och /uppdrag. Resetten är bort - Tailwind-marginalerna styr. Anledning: den gamla scoped-resetten nollställde Tailwind-marginaler inne i komponenterna, vilket gjorde det svårt att hålla ihop temat.
