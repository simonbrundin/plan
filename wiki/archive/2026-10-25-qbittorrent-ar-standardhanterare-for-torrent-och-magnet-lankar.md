---
title: qBittorrent är standardhanterare för .torrent och magnet-länkar
type: note
tags: [omarchy, workstation, torrent, xdg, projekt:infrastructure]
---

Efter installation av qbittorrent (extra/qbittorrent 5.2.4) sattes kopplingen i ~/.config/mimeapps.list via 'xdg-mime default org.qbittorrent.qBittorrent.desktop application/x-bittorrent x-scheme-handler/magnet'. desktop-filen (/usr/share/applications/org.qbittorrent.qBittorrent.desktop) deklarerar redan MimeType=application/x-bittorrent;x-scheme-handler/magnet, så kopplingen funkade även utan explicit inställning - den lagras nu explicit för att överleva en ev. paketuppdatering. Verifierat end-to-end: 'gio open fil.torrent' startar qbittorrent, och magnet-länkar måste öppnas via 'xdg-open' (använder webbläsaren) - 'gio open magnet:...' fungerar INTE, eftersom GIO behandlar magnet: som en sökväg.
