---
title: MATCHi-bokning: slot-id ändras när bokningen skapas
type: note
tags: [matchi, api, lesson, projekt:tmp]
created: 2026-11-06
updated: 2026-11-17
status: active
related: [notes/boka-enskild-kommunbokning-i-matchi-fran-bookings, notes/synk-hallbokningar-kommunens-actor-smartbook-matchi]
---

Efter att en bokning skapas i MATCHi (/facility/booking/book) får den ett NYTT slot-id. Det slot-id som skickades in (från det fria schemat) finns inte kvar och pekar inte på bokningen.

Följd: en synkronisering som sparar slot-id:t från bokningsanropet kommer att försöka flytta en bokning med ett dött slot-id.

Fix i src/go/internal/matchi/booking.go: efter bokning och efter flytt läses schemat på nytt, och den bokade cellen hittas genom att matcha bana + start + slut (findBookedSlot / findBookedSlotStartingAt). Den bokade cellen täcker hela bokningen, även om den består av flera slotter.

Viktig regel: matcha på tid, inte på slot-id, när en bokning ska återfås i MATCHis schema.

## Relaterade

- [[notes/boka-enskild-kommunbokning-i-matchi-fran-bookings]]
- [[notes/synk-hallbokningar-kommunens-actor-smartbook-matchi]]
