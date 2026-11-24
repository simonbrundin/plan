---
title: Pi Durable fungerar med MiniMax coding plan (anthropic-messages-API)
type: note
tags: [pi, durable, minimax, projekt:infrastructure]
---

Pi Durable (harness-biblioteket från earendil.com/posts/pi-durable/) kan köras direkt mot MiniMax coding plan. pi-ai har en inbyggd minimax-provider (dist/providers/minimax.js) som pratar anthropic-messages-API mot https://api.minimax.io/anthropic - samma endpoint som vår models-store.json för MiniMax. Auth sker via env-variabeln MINIMAX_API_KEY (envApiKeyAuth), ingen auth-fil behövs. Modeller som fungerar mot vår nyckel: MiniMax-M2.7, MiniMax-M2.7-highspeed, MiniMax-M3, MiniMax-M3.1-Flash. Kostnad per full demo: ca 0,9 öre (M2.7, inkl. prompt-cache) mot ca 3 öre för gpt-6.1-sol. Vår implementation: ~/repos/pi-durable-demo/model.ts centraliserar provider/modell via env (PD_PROVIDER, PD_MODEL) så både demo.ts och full.ts kan växla. Viktigt: alla tool-calls, subagent och hållbara tasks fungerade identiskt med MiniMax - inga API-specifika egenheter behövdes.
