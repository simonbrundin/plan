---
title: Chattinterface byggt i Plan - Slack borttaget
type: note
tags: [feature, agent, slack, session, projekt:plan]
---

Användaren vill inte använda Slack alls - agenten ska prata med henom i Plan. Slack-koden är nu borttagen ur agents-repot (internal/notifier/slack.go, internal/api/handlers/slack.go, SLACK_WEBHOOK_URL, webhook-ruterna) och ersatt av ett REST-API for godkannanden: GET /api/v1/approvals och POST /api/v1/approvals/:goal_id/decision. Schedulern publicerar nu forslaget i Plan (agent_runs status blocked) istallet for att skicka en notis.

Plan-side: server/api/agent/* proxar agent-API:t (konversationer, meddelanden, SSE, godkannanden) med sessionskontroll, och app/components/agent/ChatPanel.vue sitter pa mal-sidan for mal med agentflaggan.

Fyra buggar hittades när chatten testades end-to-end:
1. stripThinkBlocks byggde strangen byte for byte med string(text[i]) - det korrupterade all UTF-8 (svenska tecken blev mojibake). Fix: skriv byte for byte med bytes.Buffer.
2. MiniMax avvisade verktyg med 'parameters': null med HTTP 400 'invalid params' - och felet var tyst. Fix: tomt objektschema + felchunk skickas nu.
3. Agent-API:t lyssnade pa hardkodad port 8080 - nu stods PORT. AGENT_SCHEDULER_ENABLED och AGENT_WORKER_ENABLED kan stanga av delarna separat.
4. Nuxtrouten som streamade svaret returnerade tomt svar - Nitro gor inte forwarding nar man skriver direkt till event.node.res. Fix: returnera ReadableStream (response.body) fran handlern.

Lardom: NITRO-routes som ska stramma maste returnera en ReadableStream, inte skriva till res. Och fel i en agentkorning ska alltid skickas som ett error-chunk - en tyst fel ser ut som en agent som inte svarar.
