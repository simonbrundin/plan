---
title: Popcorn Time installerad som AppImage i ~/.local
type: note
tags: [omarchy, workstation, popcorntime, appimage, projekt:infrastructure]
---

Efter qBittorrent (se noten om mime-kopplingar) bad användaren om Popcorn Time. Valet blev inte AUR-paketet 'popcorntime' (bygger popcorn-desktop från källkod med yarn/node 2021-beroenden - risk för byggfel på node 25) utan den officiella AppImage:n från samma upstream-repo: https://github.com/popcorn-time-ru/popcorn-desktop/releases/download/v0.5.1/Popcorn-Time-0.5.1-linux64.AppImage (192 MB, NW.js/Chromium-baserad). Installationssätt: ~/.local/opt/popcorntime/Popcorn-Time.AppImage (+x), symlink ~/.local/bin/popcorntime, desktopfil ~/.local/share/applications/popcorntime.desktop (desktop-file-validate OK), ikon ~/.local/share/icons/hicolor/256x256/apps/popcorntime.png (extraherad ur AppImage via --appimage-extract). fuse2 2.9.9 finns på maskinen så AppImage-montering fungerar. Start verifierad: fönster 'Popcorn-Time' (class Popcorn-Time) dök upp via XWayland. Konfig/state hamnar i ~/.config/Popcorn-Time. VIKTIGT: sajten popcorntime.site/popcorntime-offcial.cyou är en klonad streaming-webbplats, inte appens nedladdningssida - ladda alltid från github.com/popcorn-time-ru/popcorn-desktop releases.
