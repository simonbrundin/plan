---
title: Moderposten arkiveras när .eml-bilagorna har lagts in
type: note
tags: [feature, mail, imap, projekt:simon]
---

Föregående version (se [[notes/expanderade-eml-laggs-in-som-riktiga-mejl-i-mappen-append]])
laffade in bilagorna som egna mejl, men moderposten LIGG KVAR i korgen och
listan markerade kopiorna med expanded_copy. I granskningsytan uppfattades det
som att de extraherade mejlen "hamnade under ett mejl".

Nu (src/go/imap/client.go, ExpandFolder):
- Alla bilagor har lagts in -> moderposten arkiveras (ArchiveEmail -> INBOX.Archive
  för one.com). Listan i /mail visar då bara de extraherade mejlen, som vanliga
  mejl med egen UID: de kan analyseras, godkännas, arkiveras och flyttas.
- En bilaga som inte kunde läggas in (APPEND stängs av) -> moderposten LIGGER KVAR
  och bilagan visas som en rad utan UID under den. Moderposten får aldrig försvinna
  när innehållet inte kom ut.
- Listan hämtas om när något lades in eller arkiverades (src/go/mcp/session.go),
  annars hade moderposten stuckit kvar i listan en extra gång.

Dedupe (detsamma som ovan, plus):
- En moderpost expanderas bara om bilagorna inte redan finns i mappen. Matchningen
  sker på INNEHÅLL (copyKey: avsändare + ämne + hash av brödtexten), inte bara på
  markörrubriken X-Pappa-Expanded-From. Skäl: markören innehåller moderpostens UID,
  och UID ändras om moderposten flyttas ut ur mappen och tillbaka. Då matchar
  markören inte längre och samma bilaga läggs in igen. Som sista utväg jämförs
  ämne + datum med övriga mejl i mappen.

Lärdomar från live-körning mot INBOX.Simon:
- one.com lägger till Received-rubriker på APPENDade mejl. Jämförelsen får därför
  inte bygga på hela rubrikblocket - använd ämne, avsändare och brödtext.
- Servern returnerar ämnet avkodat för UTF-8 men rått för Windows-1252. Båda sidor
  ska därför avkodas innan jämförelse (decodeMIMEHeader).
- Ett fel i dedupen gjorde att 15 extra kopior lades in i korgen. De rensades bort
  till INBOX.Trash.Deleted Messages. Korgen har nu 23 mejl: 4 original + 15 kopior
  från moderpost uid 43 + 4 kopior från moderpost uid 59. Moderposterna ligger i
  INBOX.Archive.
