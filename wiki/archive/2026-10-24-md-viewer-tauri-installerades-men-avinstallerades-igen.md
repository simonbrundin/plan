---
title: MD Viewer (Tauri) installerades men avinstallerades igen
type: note
tags: [tool, desktop, install, projekt:infrastructure]
---

Installerade MD Viewer v0.1.9 (CreateChance/mdviewer) på Omarchy-datorn, men användaren avinstallerade den direkt - "fungerade jätte dåligt". Lärdom: gissa inte att en ny desktop-app passar användaren, fråga först vilken han redan använder (för markdown är det typora, vilket också är registrerad som standardhanterare för text/markdown).

Tekniskt: uppdragsgivaren exponerar inget Arch-paket (bara .deb/.rpm/AppImage) och maskinen har ingen sudo utan lösenord, så .deb-paketet packades upp manuellt med ar+tar till ~/.local/bin, ~/.local/share/applications och ~/.local/share/icons/hicolor. Beroenden (webkit2gtk-4.1, gtk3) fanns redan. Avinstalleringen innebar att ta bort de tre platserna, köra update-desktop-database och radera de två mdviewer-raderna ur ~/.config/mimeapps.list - annars pekade standardschemat kvar en daterad desktop-fil.
