---
title: Rantebetalningsmejl hamnar i INBOX.Deleted Messages - och arkivet matchades aldrig
type: note
tags: [kameo, mail, imap, bug, projekt:pappa]
created: 2026-11-16
updated: 2026-11-17
status: active
related: [notes/kameo-ranteutbetalning-redan-utbetald-samlingsuttag-pa-47-29-kr]
---

Tvinsikter om var rantebetalningsmejlen ligger, och en bugg i filtreringen av mappar.

1. De senaste rantebetalningsmejlen hamnade i INBOX.Deleted Messages, inte i INBOX.Kameo.*. En sökning som bara tittade pa Kameo-mapparna hittade alltsa aldrig det aktuella. Fixen: interestFolders tar med INBOX, mappar med 'kameo' i namnet, och mappar vars sista del ar 'Deleted Messages' eller 'Arkiv'/'Archive'.

2. BUGG: filtret var /deleted messages|^archive$/i. Regexet med ^...$ krävde att HELA mappnamnet var 'archive', sa 'INBOX.Archive' matchade inte. Arkiverade rantebetalningsmejl kunde alltsa aldrig hittas. Fixen: matcha pa mappens SISTA DEL (folder.split('.').pop()) sa INBOX.Archive och INBOX.Kameo.Arkiv traffar samma regel. Testet 'rekommenderade mappar omfattar Kameo, raderade och arkivet' fangade detta.

3. one.com har 422 mappar i Michael Brundins konto - en checklista med filtertextfalt och knapparna Rekommenderade / Lagg till synliga / Ingen fungerar bra. Knappen 'Rekommenderade' med filter aktivt galler de rekommenderade mapparna UNDER filtret, inte alla.

Varfor detta matt: en anvandare som vill ha kontroll over vad som sokes maste kunna välja mappar, men standardvalet ska fortfarande vara de mappar som normalt innehaller rantebetalningar - annars missar man de aktuella.

## Relaterade

- [[notes/kameo-ranteutbetalning-redan-utbetald-samlingsuttag-pa-47-29-kr]]

