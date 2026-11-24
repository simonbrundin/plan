---
title: Popcorn Time: 'remote movies API failed to respond' - orsak och fix
type: note
tags: [popcorntime, omarchy, bug, xdg, projekt:infrastructure]
created: 2026-10-25
updated: 2026-11-17
status: active
related: [notes/popcorn-time-svenskt-ljud-finns-inte-i-katalogen-anvand-torrentsamling-torrent-collection]
---

Symtom: Popcorn Time 0.5.1 (AppImage) startar men listan är tom och felet 'The remote movies API failed to respond' visas. Orsak: releasepaketet ships node_modules/butter-settings-popcorntime.app version 0.0.10 där providers-uri har TOM apiURL ('MovieApi?apiURL='). De gamla byggda API-värdarna (movies-v2.api-fetch.sh/.am/.website) är sedan flera år DÖDA (NXDOMAIN, DoH-status 2/3) och yts.mx likaså. Appen hämtar normalt API-listan via DHT-nyckeln a4a9ad29e303e137ecb995c50a4e104b3e8f72e5, men ingen seedar längre - så auto-uppdateringen ger inga servrar. Fix (2026-10-24): skrev manuella servrar som settings-dokument i den nedb-filen ~/.config/Popcorn-Time/Default/data/settings.db (en JSON-rad per dokument): customMoviesServer/customSeriesServer/customAnimeServer = 'https://fusme.link/,https://jfper.link/,https://uxert.link/,https://yrkde.link/'. Appkoden (database.js -> App.Providers.updateConnection) läser dessa nycklar vid start och sätter provider.setApiUrls(). Viktigt: värdena måste sluta på '/'. Verifierat via CDP (node + global WebSocket mot NW.js --remote-debugging-port=9222): Movies (Genre All 115426), Series (21719) och Anime renderar nu kataloger. Framtida API-doditions: dessa .link-domäner är tredjepartsservrar och kan dö - kontrollera med curl -s -o /dev/null -w '%{http_code}' <url>/movies/1?limit=1.

## Relaterade

- [[notes/popcorn-time-svenskt-ljud-finns-inte-i-katalogen-anvand-torrentsamling-torrent-collection]]

