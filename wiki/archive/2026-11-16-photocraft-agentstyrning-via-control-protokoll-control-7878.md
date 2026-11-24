---
title: PhotoCraft agentstyrning via control-protokoll (--control 7878)
type: note
tags: [photocraft, mcp, automatisering, omarchy, projekt:simon]
---

För att låta en AI-agent redigera BILDEN SOM ÄR ÖPPEN i PhotoCraft (inte headless) måste appen startas med --control och en tokenfil, och MCP-klienten köras i bridge-läge.

Kommando som fungerade:
setsid photocraft --control 7878 --control-token-file ~/.config/photocraft/agent-control.token --automation-read-root ~/photos --automation-write-root ~/Pictures/PhotoCraft/automation <fil>

Varför detta väg: preferences.json (integrations.connectAgents/controlPort) läses bara vid start, och utan --control-flaggan sätts inga automation-rötter -> 'write authority is absent'. Med --control är rötterna explicita. Tokenfilen gör att token inte ändras mellan omstarter (token i stderr annars).

Fällor hittade:
- instance.json i ~/.config/photocraft är INTE control-anslutningen utan single-instance-IPC (svarar 'not taken'). Kontrollporten är en helt annan listener.
- Kontrollprotokollet använder challenge-auth (auth.challenge -> HMAC-SHA256 med token som nyckel, label 'photocraft-control-v1 server/client'); efter challenge accepteras ENDA auth {proof}, inte token-auth.
- engine.execute vägrar kommandon med absoluta/amienta sökvägar ('uses ambient filesystem paths and is disabled'). Använd capability-scoped MCP-verktyg (doc_export/doc_save) eller relativa sökvägar under write root.
- ui.screenshot/frame-beroende kommandon time:a ut om fönstret inte presenterar frames (t.ex. ligger på en annan workspace/occluderad). Fokus fönstret först.
- 'photocraft-cli mcp --bridge' i pi: pi-sessionen cachar mcp.json, så en nydagd server syns först efter /reload eller ny session.

Kommandot att sätta upp MCP i pi: photocraft -> photocraft-cli mcp --bridge 127.0.0.1:7878 --control-token-file ~/.config/photocraft/agent-control.token ; photocraft-headless -> vanlig headless med rötter.
