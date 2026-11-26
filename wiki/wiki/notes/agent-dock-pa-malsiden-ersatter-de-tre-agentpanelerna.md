---
title: Agent-dock på målsiden ersätter de tre agentpanelerna
type: note
tags: [agent, granssnitt, ui, feature, projekt:plan]
created: 2026-11-26
updated: 2026-11-26
status: active
---

2026-11-26: På målsidan (/goal/:id) låg agentchatt, körningshistorik och checkpoint-plan i samma lodräta flöde som målets innehåll. Chatten var 24rem hög och körningshistoriken obegränsad, så undermålen trängdes nedanför och det krävdes mycket scrolling för att nå dem.

Ändring: agentdelen flyttades till en egen kolumn (AgentDock, src/nuxt/app/components/agent/Dock.vue) med flikarna Chatt / Körning / Plan - bara en del syns åt gången. Panelen ligger FAST höger i fönstret: på lg+ ligger den ovanför kalenderspalten (24rem/26rem) så att målets innehåll behåller full bredd, på mindre skärmar lägger den sig över sidan med en bakgrund. Valet (öppen/stängd) sparas i localStorage under plan.agentDockOpen; standard är öppen på lg+.

Varför fast och inte i flödet: appen har redan en permanent högerspalt (kalendern). En dock i flödet skulle ha trängt ut målkolumnen till ~575px på en 1440px-skärm, eftersom huvudytan bara är 1024px bred (kalendern tar 416px).

Andra förbättringar i samma svep: chatten har Enter för att skicka (Shift+Enter för ny rad), tidsstämpel och roll per meddelande, verktygsanrop ihopfällda, knapp för att hoppa till nya meddelanden; körningshistoriken är en kompakt tidslinje med svenska etiketter, senaste först, payload bakad Visa detaljer och Visa fler; tangenten mellanslag + v visar/döljer agentpanelen.

Lärdom: i den här appen är skärmbredd en resurs - kontrollera alltid vad som redan ligger i sidan (kalenderspalt, sidebar) innan en ny kolumn läggs till.
