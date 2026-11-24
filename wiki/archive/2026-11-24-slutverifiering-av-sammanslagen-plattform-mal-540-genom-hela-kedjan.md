---
title: Slutverifiering av sammanslagen plattform: mål #540 genom hela kedjan
type: note
tags: [plan, agent, verifiering, e2e, human-gate, kubernetes, wiki, projekt:simon]
---

Efter sammanslagningen (mål 520) gick hela kedjan igenom i dev med mål #540:

AgentRun #64 -> Job agent-worker-run-64 -> godkännande #51 (godkändes via
gränssnittet) -> agent/goal-540-run-64 publicerad, huvudgrennen orörd.

Händelsekronologin: run_started, approval_requested, approval_decided,
phase_changed, artifact, verification, artifact, run_completed.

Vad som gjorde att det fungerade:
- Scopade API-nycklar. Agentens nyckel kan inte besluta om ett godkännande
  (HTTP 403) - human gate är en teknisk gräns, inte en instruktion.
- Agent-API:t kräver AGENT_API_KEY, och Nuxt skickar den server-side.
- Godkännandena ligger i Plan, så UI:t läser och beslutar direkt där. Den
  mänskliga sessionen är det som godkänner - aldrig agentens nyckel.
- Varje körning får sin egen branch (agent/goal-<mål>-run-<körning>), så
  upprepade försök inte krockar.

Lärdom: när två flöden (den gamla agent-UI:ns filbaserade godkännanden och
den nya Plan-ägda human gate) skaSamexistera, välj en ägare. Att ha båda gjorde
det oklart vilken väg beslutet tog. Lösningen var att låta Plan äga
godkännandena och låta agenten bara observera.

Kommits i plan-repot: sju avgränsade commits (agent-backend, Plan-modell,
behörigheter, human gate i UI, miljöer, wikin, gofmt).
