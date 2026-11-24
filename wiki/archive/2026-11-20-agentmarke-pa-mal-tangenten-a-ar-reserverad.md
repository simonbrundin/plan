---
title: Agentmärke på mål - tangenten a är reserverad
type: note
tags: [feature, goals, tangent, projekt:plan]
---

Mål har nu en agent-flagga (goals.agent i databasen) som visas med AgentBadge (blått när på) och växlas med leader + a (mellanslag + a) på mållistan och målsidan.

Varför inte plain 'a'? Tangenten 'a' används redan av tangentbevakare för att flytta markören vid titelredigering, så en egen tangent skulle krocka. Leader-sekvensen (leader + a) är registrerad som sidkommando via registerPageCommands och visas i leader-modalens kommandolista.

Undo: z ångrar agent-växlingen (undoAgent i useGoalUndo).
