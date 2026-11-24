---
title: Beslut: Pi Durable sätts på hyllan för tillfället
type: note
tags: [beslut, pi, durable, projekt:infrastructure]
---

Efter att ha byggt och verifierat en fungerande Pi Durable-demo (~/repos/pi-durable-demo, demo.ts + full.ts, körs med MiniMax via PD_PROVIDER/PD_MODEL) beslöt Simon att hoppa över Pi Durable heltills vidare. Inget installerades globalt: pi list oförändrad, inga settings.json-ändringar, inga bakgrundsprocesser. Demon ligger kvar som referens (113 MB, eget git-repo, inga remote) och kan raderas med rm -rf ~/repos/pi-durable-demo. Inlärningen är värd att behålla eftersom den är dyrt att härleda: Pi Durable är ett npm-bibliotek, inte en tjänst - det krävs bara en mapp med package.json + node_modules och en SQLite-fil; verktyg returnerar { output: [...] } (inte { content: [...] }); defineDoc kräver checkpointWhen; tx.createTask utanför konversationsbundet commit kräver conversationId; en task-fas som gör fs-I/O med fasens abortSignal hänger om invokationen avbryts, så vänta på tasken innan harness.close(). Mer generellt: det vanligaste sättet att använda Pi Durable i praktiken är paketet pi-agents (npm, peer deps ^1.4.0) som är en bro från Pi till pi-durable, plus Cloudflare PiHarness/Vercel durable-harness för agenttjänster.
