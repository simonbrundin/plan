---
title: Artikeln om agentic software factory - vad som är relevant för agentsystemet
type: note
tags: [artikel, arkitektur, wiki, governance, projekt:agents]
---

Källa: https://levelup.gitconnected.com/build-an-agentic-software-factory-f995f79221ef (Akshay Kokane, del 3 i AI-Native SWE-serien). Läst via sökning (Medium blockade direkt-fetch).

## Artikelns fem byggblock
1. Planering/dekomponering - krav till körbara tasks med beroenden och acceptance criteria.
2. Persistent engineering memory - ADR:er, auktoritetskedja, inte bara mer dokument.
3. Agentkoordinering - strukturerad uppgiftsöverlämning (uppgift, antaganden, beslut, beroenden, criteria, olösta frågor), inte prosa.
4. Oberoende verifiering - klart = bevis, inte agentens säkerhet.
5. Governance - vem godkänner, när auto-deploy, vilka system agenten får röra, spend-gränner, eskalering.

## Bedömning
Ramverket är bra men ooperationaliserat (inga mätetal, ingen standard). Multi-agent-parallelism och auto-merge bör skjutas upp. Vi har redan byggt en factory: goals/Plan = block 1, wikictl = block 2, orchestrator+checkpoints = block 3, gates = block 4, approvals+Slack = block 5.

## Tre konkreta förbättringar (prioriterade, ej implementerade)
A. Auktoritetskedja i wikin: lägg till 'supersedes'/'superseded_by'-länk mellan sidor samt 'reconsider_when' (mätning + ägare) på beslut. Lint-regel: deprecated-sida måste peka på ersättare. Utan detta kan en agent bara se att något är deprecated, inte vilket beslut som gäller.
B. Beslutversion i checkpoint-kontraktet: checkpoint-planneren ska skriva in vilket beslut (id, status, datum) som konsulterades, så adversarial-gate kan upptäcka att beslutet ersatts under arbetet.
C. Governance som deklarativ fil: policy.yaml i repot med tillåtna kommandon, tillåtna miljöer, krav på human-gate - deterministisk enforcement istället för prompt-instruktioner.

## Redan starkare än artikeln
Vår sensor-pipeline (wiki som sensor, goals som intake) ligger före 'start with the sensor'-rörelsen i The New Stack, och vår verifiering är konkretare än artikelns 'evidence'-vagueness.
