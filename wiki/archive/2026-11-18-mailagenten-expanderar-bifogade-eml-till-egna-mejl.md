---
title: Mailagenten expanderar bifogade .eml till egna mejl
type: note
tags: [feature, mail, imap, projekt:pappa]
---

När inkorgen hämtas (/mail -> "Hämta mejl") expanderas nu bifogade .eml-filer
och message/rfc822-paket till egna rader i listan, direkt under moderposten.

Varför: ett inkorgsmejl (t.ex. en Outlook-paketforward med 15 vidarebefordrade
mejl) annonserade bara sig som ett tomt mejl. Innehållet fanns i bilagorna
och syntes aldrig.

Hur: src/go/imap/expand.go går igenom MIME-delen rekursivt efter
message/rfc822 och bilagor med filnamnet .eml, parsar varje bilaga som ett
mejl och returnerar dem som EmailContext med FromAttachment/ParentUID/Depth.
ListEndpoints svaret får ett eget id per expanderat mejl ("x<parent>-<depth>-<nr>").

Viktiga lärdomar:
- one.com/Dovecot rapporterar MIME-typer i LOWER case ("message/rfc822"), så
  jämförelser måste vara case-insensitive (EqualFold). En exakt jämförelse
  mot "MESSAGE" tystgick förbi alla bilagor.
- Ett expanderat mejl finns inte i brevlådan och har ingen UID. Det kan därför
  inte arkiveras, flyttas eller läsas från servern - åtgärder och AI-analys
  filtrerar bort dem (mailboxEmails i src/go/mcp/session.go).
- Expansjonen sker i två steg: brevstrukturen (BODYSTRUCTURE) avgör vilka
  mejl som har .eml-bilagor, sedan hämtas bara deras hela brödtext.

Verifierat mot riktig post: INBOX.Simon, uid 43 (15 bilagor) -> 20 rader i
listan via Nuxt-proxyn på localhost:3000.
