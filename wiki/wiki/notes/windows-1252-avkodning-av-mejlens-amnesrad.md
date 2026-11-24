---
title: Windows-1252-avkodning av mejlens ämnesrad
type: note
tags: [bug, go, imap, mime, projekt:pappa]
created: 2026-11-20
updated: 2026-11-20
status: active
---

2026-11-19: Ämnen på Outlook-mejl visades rått MIME-kodat i pappa, t.ex. "=?Windows-1252?Q?VB:_Senior_Technical_Project?=". Orsaken var att decodeHeader i src/go/imap/client.go var en no-op ("Simple decoder - for production use go-message package").

Varför: go-imap avkodar bara UTF-8 och US-ASCII via sin mime.WordDecoder. Windows-1252 är det teckenkodning Outlook använder för svenska ämnen, så avkodningen misslyckades och headern visades rått.

Fix: en egen mime.WordDecoder med en CharsetReader som hanterar windows-1252, iso-8859-1 och iso-8859-15 via golang.org/x/text/encoding/charmap (x/text var redan ett dependency). Headern avkodas ord för ord - en kodad word får inte innehålla mellanslag, så delningen är säker, och det bevarar mellanslag mellan två kodade ord (MIME säger att sådana ska ignoreras).

Generell regel: när ett bibliotek bara stöder UTF-8/US-ASCII och mejl från Outlook kräver windows-1252, är rätt ingångspunkt en egen CharsetReader - inte att lappa per anropsplats.
