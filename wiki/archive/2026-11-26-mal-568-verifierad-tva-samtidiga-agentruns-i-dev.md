---
title: Mål 568 verifierad: två samtidiga AgentRuns i dev
type: note
tags: [agent, kubernetes, job, verifiering, projekt:plan]
---

Verifiering 2026-11-26 i dev-klustret (kind-dev, agent-api).

Genomförd: skapade två köade AgentRuns (86 för mål #575, 87 för mål #576, profil 'life', olika mål). Jobbcontrollern skapade båda jobben i samma reconcile-pass 23:07:40 (agent-worker-run-86 och -87), båda pods Running samtidigt (active=1) och båda körningarna slutfördes som succeeded.

Varför detta räcker: målet gällde att jobbcontrollern ska hantera flera samtidiga AgentRuns. Den faktiska parallelliteten är strukturell - varje körning får ett eget Kubernetes Job - så beviset är att flera jobb levde samtidigt, inte att någon enskild körning var snabbare.

Viktig avgränsning: implementationen serialiserar medvetet samma mål (HasActiveJobForGoal) - två körningar mot samma mål körs inte parallellt eftersom de skulle skriva över varandra i samma repo. Parallelliteten gäller alltså olika mål, upp till AGENT_JOB_MAX_CONCURRENT (standard 4).

Testmetod som fungerar: skapa goals via API, skapa runs via POST /agent-runs med agent_profile=life (ingen repo-kloning, ingen PR), bevaka kubectl get jobs -l app=agent-worker efter två aktiva jobb, och städa sedan med DELETE /goals/:id plus kubectl delete job.

Resterande: körningarna 86 och 87 finns kvar som historik i Plan - det finns ingen DELETE-endpoint för agent-runs.
