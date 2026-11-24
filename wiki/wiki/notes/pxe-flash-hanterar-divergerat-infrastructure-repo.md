---
title: PXE-flash hanterar divergerat infrastructure-repo
type: note
tags: [bug, pxe, git, projekt:infrastructure]
created: 2026-11-20
updated: 2026-11-20
status: active
---

simon flash misslyckades på PXE-servern eftersom ~/infrastructure låg 978 commits före och 1072 efter origin/main; git pull vägrade välja merge-strategi. Repo:t var rent men stale. Backup-branchen backup/pxe-before-sync-20261120-1528 skapades och checkouten synkades till origin/main. pxe/flash.nu uppdaterades till fetch origin/main + kontroll av lokala ändringar + backup-bransch + reset till origin/main vid divergens, och ändringen committades som 14d37f76 och pushades.
