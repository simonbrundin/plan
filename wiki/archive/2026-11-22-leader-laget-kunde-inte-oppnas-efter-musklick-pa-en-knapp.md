---
title: Leader-läget kunde inte öppnas efter musklick på en knapp
type: note
tags: [bug, leader, keyboard, focus, projekt:plan]
---

Symtom: på /now (och andra sidor) verkade leader-kommandona inte fungera - mellanslag öppnade inte leader-modalen.

Varför: appar ritar fokusringen med :focus-visible, så en knapp som klickats med musen ser ut att sakna fokus - men är fortfarande aktiveElement. Leader-läget öppnas bara med mellanslag om fokus inte ligger på en knapp (så att en fokuserad knapp kan aktiveras med mellanslag). Efter ett musklick på t.ex. filtren på /now, en radknapp eller en ikonknapp låg fokus kvar på knappen, så mellanslag aktiverade knappen i stället för att starta leader-sekvensen. På /now blev det extra tydligt eftersom man måste klicka 'Alla' (filter 'Nu' är ofta tomt) och då försvann tangenterna.

Fix: ny plugin src/nuxt/app/plugins/mouse-focus.client.ts. Den lyssnar på 'click' i CAPTURE-fasen (många knappar använder @click.stop, så bubblingsfasen når inte document) och gör button.blur() i en setTimeout(0), alltså efter knappens egen klick-hanterare. Endast musklick påverkas: event.detail === 0 (tangentbordsklick) lämnas orörda, så knappar som nåtts med Tab fortfarande kan aktiveras med mellanslag.

Verifierat: bun run build, bun run test (89 scenarier), bun test test/unit (216 pass), samt headless chromium mot den byggda appen: efter musklick på filterknapp/ikonknapp öppnar Space leader-läget på /now och /goals, leader-kommandona (a, n, d) fungerar, vanliga tangenter (d) fungerar oförändrat, och efter Tab-aktivering öppnar Space fortfarande knappen (inte leader-läget).
