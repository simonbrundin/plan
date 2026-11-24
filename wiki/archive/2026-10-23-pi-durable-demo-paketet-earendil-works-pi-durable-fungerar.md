---
title: Pi Durable-demo (paketet @earendil-works/pi-durable fungerar)
type: note
tags: [pi, durable, agent, projekt:infrastructure]
---

Pi Durable (Earendils experimentella harness, https://earendil.com/posts/pi-durable/) finns publicerad på npm som @earendil-works/pi-durable@1.4.0 och fungerar direkt med Node 22.19+/25 och TypeScript (node demo.ts räcker, ingen byggsteg). Vår fungerande demo ligger i ~/repos/pi-durable-demo: tre lägen (run, crash+resume, fork) med SQLite-lagring, CodingTools och NodeExecutionEnv. Viktiga lärdomar: (1) API-kedjan är Harness.open -> harness.root(ctx, {agent}) -> conversation.submit -> submission.wait -> tx.entry(AssistantEntry, answer); (2) pi-ai läser OPENAI_API_KEY ur process.env, så pi:s egen ~/.pi/agent/auth.json måste läsas in och sättas som env-variabel; (3) node --experimental-strip-types fungerar direkt, men package.json måste ha "type": "module" annars körs .ts som CJS; (4) crash/resume kräver att processen dörs under en pågående modellturn - enbart att stänga processen räcker inte för att få en avbruten turn att återupptas.
