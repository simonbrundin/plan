---
title: Expanderade .eml läggs in som riktiga mejl i mappen (APPEND)
type: note
tags: [feature, mail, imap, projekt:pappa]
created: 2026-11-18
updated: 2026-11-19
status: active
related: [notes/mailagenten-expanderar-bifogade-eml-till-egna-mejl, notes/moderposten-arkiveras-nar-eml-bilagorna-har-lagts-in]
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

Så undviks dubbletter: en moderpost expanderas bara om bilagorna inte redan
finns i mappen. Matchningen sker på innehåll (ämne, avsändare, brödtext) och
inte bara på markörrubriken - se
[[notes/moderposten-arkiveras-nar-eml-bilagorna-har-lagts-in]]. Moderposten
arkiveras när bilagorna är uttagna, så att listan visar de extraherade mejlen
som vanliga mejl.

Fallback: om servern inte tillåter APPEND läggs bilagorna inte in, utan
visas som rader utan egen UID i listan (expanded=true). De kan då inte
analyseras eller arkiveras.

Verifierat mot INBOX.Simon: moderpost uid 43 med 15 bilagor -> 15 nya mejl
(uid 44-58) i mappen, markerade som expanded_copy i listan. Analys av
limit=3 returnerar analys av uid 56-58, alltså kopiorna.
