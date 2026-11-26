---
title: Varfor src har tva cmd-kataloger: Go internal-regeln sparrar gemensam cmd
type: note
tags: [go, struktur, plan, projekt:plan]
created: 2026-11-26
updated: 2026-11-26
status: active
---

2026-11-26 flyttades binarerna till cmd/: src/api/cmd/{api,rrt} och src/agent/cmd/{agent,agentworker,wikictl}. Intuitionen 'en gemensam src/cmd/' provades forst - och failade: Go internal-regeln gor plan/agent/internal/... oimporterbar utanfor plan/agent-tradet, sa en main under plan/cmd kan inte n paketen. Alternativet (modulrot-internal/) skulle lagt av kompilatorn tvingade api<->agent-isoleringen - samma disciplin som plan-agent.yaml-kontraktet - pa konvention. Darfor: tva cmd-roter, en per deployable. agent/main.go lades till agent/cmd/agent (main-paket pa paketroten var icke-idiomatiskt) och api/cmd/server doptes om till api/cmd/api. Dockerfiler, agent-go.yml och docs uppdaterade. rrt (api/cmd/rrt) ar en rrule-scratch utan produktkoppling - recurrence_test.go tacker beteendet.
