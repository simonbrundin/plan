---
title: Expanderade .eml läggs in som riktiga mejl i mappen (APPEND)
type: note
tags: [feature, mail, imap, projekt:pappa]
---

Förra versionen av .eml-expansionen (se [[notes/mailagenten-expanderar-bifogade-eml-till-egna-mejl]])
visade bilagorna bara i listan, i minnet. De expanderade mejlen hade ingen
UID och kunde därför inte analyseras, godkännas, arkiveras eller flyttas.

Nu läggs varje bifogad .eml in som ett eget mejl i samma mapp via IMAP APPEND
(src/go/imap/client.go, ExpandFolder/appendExpanded). Kopian märks med
X-Pappa-Expanded-From: <moderpostens UID>.

Varför APPEND och inte enbart minnes-expansion: expanderade mejl måste gå
igenom hela kedjan - "Processa mejl" ska analysera dem, och arkivera/svara
kräver en UID som finns i brevlådan. Ett riktigt mejl i mappen får allt
detta gratis, utan att analys- och beslutsmodellen behöver ändras.

Så undviks dubbletter: en moderpost expanderas bara om mappen saknar
kopior av den (UID SEARCH HEADER X-Pappa-Expanded-From "<uid>"). Därför
expanderas moderposten bara en gång - vid första listningen - och
kopiorna syns som vanliga mejl vid alla efterföljande listningar.

Fallback: om servern inte tillåter APPEND läggs bilagorna inte in, utan
visas som rader utan egen UID i listan (expanded=true). De kan då inte
analyseras eller arkiveras.

Verifierat mot INBOX.Simon: moderpost uid 43 med 15 bilagor -> 15 nya mejl
(uid 44-58) i mappen, markerade som expanded_copy i listan. Analys av
limit=3 returnerar analys av uid 56-58, alltså kopiorna.
