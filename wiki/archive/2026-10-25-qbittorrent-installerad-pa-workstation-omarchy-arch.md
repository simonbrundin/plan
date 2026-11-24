---
title: qBittorrent installerad på workstation (Omarchy/Arch)
type: note
tags: [omarchy, workstation, torrent, projekt:infrastructure]
---

Användaren bad om en torrent-klient på datorn (Omarchy/Arch, x86_64). Val: qBittorrent 5.2.4 från extra (pacman -S qbittorrent) - Qt6-GUI, libtorrent-rasterbar, och även qbittorrent-nox finns i samma paket för headless/daemon-läge. Inget var installerat tidigare och ingen flatpak finns på maskinen. Om WebUI-varianter önskas kan qbittorrent-nox köras som systemd --user-tjänst (kräver egen setup med auth-config).
