---
title: Chattkonversationer sparas per mål
type: note
tags: [feature, agent, chatt, mål, projekt:plan]
created: 2026-11-27
updated: 2026-11-27
status: active
---

VARFÖR: Chatten med agenten såg ut att den inte sparades. Meddelandena fanns
i agentens databas (plan-agent-db), men det fanns ingen koppling mål →
konversation som gick att hitta tillbaka. Enda länken var
agent_runs.conversation_id i Plan-databasen, och den sattes bara när målet
redan hade en agentkörning (saveConversationId returnerar tidigt annars).
Följden: ett mål utan körning fick en NY konversation varje gång chatten
öppnades, och de gamla samtalen blev spök i databasen. I Neon dev hade 14 av
de 15 senaste agent_runs conversation_id = NULL.

HUR: conversations-tabellen i agentens databas fick kolumnen goal_id
(BIGINT, Plan-målets id) med index. POST /api/v1/conversations tar goal_id
(valfritt) och sparar det; GET /api/v1/conversations?goal_id=<id> listar
samtalen för ett mål. Gränssnittet laddar det senaste samtalet för målet
när målsidan öppnas, listar tidigare chattar i en rullmeny och erbjuder
"Ny chatt". Både Go- och Rust-agenten är uppdaterade (samma JSON), och
swagger-filen är synkad med hand (swag-genereringen gav för stor diff).

LÄRDOM: agent_runs.conversation_id är fel ställe att hålla kopplingen - en
körning är inte samma sak som ett samtal. En chatt kan finnas utan att någon
körning startats, så länken måste ligga på samtalet. Meddelandena bor i
agentens databas, så det är där indexet ska ligga.

Verifierat i kind-dev 2026-11-27: mål 540, samtal skapat via Nuxt-proxy,
listat via GET ?goal_id=540, migrationen körd mot befintlig agentdatabas.
