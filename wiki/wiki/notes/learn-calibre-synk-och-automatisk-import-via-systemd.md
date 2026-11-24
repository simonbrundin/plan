---
title: Learn: Calibre-synk och automatisk import via systemd
type: note
tags: [calibre, automatik, systemd, import, projekt:simon]
created: 2026-11-03
updated: 2026-11-03
status: active
related: [notes/learn-kunskapsbasen-kan-lagras-i-postgresql]
---

Användaren bad om att få sina Calibre-böcker in i Learn och att de skulle importeras automatiskt.

Loesning i Go (repot github.com/simon/repo-template):
- internal/calibre läser Calibres metadata.db via modernc.org/sqlite. Öppnas skrivskyddat med file:-URI och mode=ro plus busy_timeout, så Calibre kan köra samtidigt. Calibres timestamp-kolumn är en datumsträng, inte ett tal - toTime() hanterar båda.
- Formatval: EPUB före ORIGINAL_EPUB, annars AZW3/MOBI/PDF. Sökvägen byggs som bibliotek/<books.path>/<data.name><extension>.
- internal/learn/sync.go: ny bok importeras, redan importerad hoppas över, ändrad fil (mtime > imported_at) importeras om via DeleteBook, misslyckad import registreras i import_failures-tabellen och prövas inte om förrän filen ändras, filer över --max-file-size hoppas över.
- learn sync med --watch, --library, --refresh, --dry-run, --limit, --max-file-size.
- systemd: learn-sync.service + .timer i src/go/systemd/, installerade till ~/.config/systemd/user, kör var tionde minut med flock så två synkar inte kör samtidigt.

Viktiga lärdomar:
1. Skrivskyddat läsning av en extern databas är inte svårt - file:path?mode=ro funkar i modernc.org/sqlite. Det gör att Calibre kan vara igång.
2. Storleksgräns är nödvändig. Calibre-biblioteket innehöll e-puber pa 1,5 GB, 636 MB och 236 MB - skannade manualer utan text. Forsta synken tog 11 min 42 s mestadels I/O. Med --max-file-size 150 MB hoppas de over och synken tar 8 ms nar allt redan ar synkat.
3. Misslyckande maste sparas, inte loggas varje gang. Utan det fyllde journalen var tionde minut med samma fem fel.
4. Ett definition-filtrag i internal/text (definitionen far inte upprepa termen mer an två ganger) tog bort brusiga definitioner i skannade manualer, tex FARA med en run-on-text. Bara 7 koncept av 4035 var drabbade, sa de fem bokerna importerades om i stallet for hela biblioteket.

Resultat pa maskinen: 69 av 78 e-puber importerade (4 for stora, 5 utan text), 4035 koncept, i ~/.learn/learn.db som API:t och SPA:t laser.

## Relaterade

- [[notes/learn-kunskapsbasen-kan-lagras-i-postgresql]]

