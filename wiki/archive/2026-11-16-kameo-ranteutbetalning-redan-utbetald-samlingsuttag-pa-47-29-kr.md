---
title: Kameo ranteutbetalning redan utbetald - samlingsuttag pa 47,29 kr
type: note
tags: [kameo, rante, dubbeluttag, projekt:pappa]
---

Kameo-flodet for ranteutbetalningar hittade fyra aktuella rantebetalningsmejl i INBOX.Deleted Messages (lån 5041, 5116, 5066, 5012 - 11,67 + 5,83 + 20,42 + 9,37 = 47,29 kr, transaktionsdatum 2026-11-16). Samtidigt fanns EN uttagsbegaran pa 47,29 kr (id 250526, status 'Uttagsbegaran behandlas', is_locked true) och Kameos tillgangliga belopp var 0,0000.

Varfor detta matt: rantebetalningsmejlet innehaller inte beloppet - det maste hamtas ur Kameos transaktionslista genom att matcha transaktionsdatumet mot mejldatumet. Och de senaste rantebetalningsmejlen hamnar inte i INBOX.Kameo.* utan i INBOX.Deleted Messages, sa en sökning som bara tittar pa Kameo-mapparna missar alltid de aktuella.

Den pausende begaran pa 47,29 ar ett samlingsuttag for hela dagens ranta, inte fyra enskilda uttag. Michael har alltsa redan hanterat rantan manuellt idag. Ett dubbelbetalningsskydd som bara jamar ett enskilt rantebelopp mot en pausande begaran (11,67 mot 11,67) hade inte upptackt detta. Regeln ar nu: en pausande begaran som TACKER beloppet (belopp >= rantebeloppet) blockerar uttaget, med forklaringen att pengarna redan ar pa vag ut.

Viktig larm: den 2026-11-16 behovde den aven anpassad regex - bara en exakt jamtarelse ar for liten.
