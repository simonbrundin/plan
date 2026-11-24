---
title: Interaktiv målchatt behöver säkra Plan-verktyg
type: note
tags: [bug, agent, chat, projekt:agents]
---

Interaktiv målchatt försökte tidigare använda bash i agent-api-containern och loopade tills maxgränsen. Chatten saknade målkontext och kunde inte skriva undermål till Plan. Åtgärd i Plan: chatten får goal_context, inga shellverktyg, och ett enda create_subgoal-verktyg som skapar och kopplar ett undermål via Plan API när användaren uttryckligen ber om det. Kodändringar görs fortfarande i isolerade AgentRuns.
