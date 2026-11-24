---
title: Sammanslagning av agents-repot in i plan: vad som flyttades och vad som medvetet inte gjorde det
type: note
tags: [plan, agent, migrering, repo, sammanläggning, ci, kubernetes, projekt:simon]
---

Fas 5 i mål 497 (mål 520) slutfördes 2026-11-24. agents-repot och plan-repot är nu ett repo.

Vad som flyttades:
- agents/src/go -> plan/src/agent (utvecklat vidare, inte bara kopierat)
- agents/wiki -> plan/wiki. ~/wiki pekar nu på ~/repos/plan/wiki, och wikictl byggs från plan/src/agent/cmd/wikictl. Ägandet av wiken flyttades i AGENTS.md.
- agents/.github/workflows/go.yml -> plan/.github/workflows/agent-go.yml (paths omdöpta till src/agent).
- agents/environments -> plan/environments/base/agent* (dev och prod delar samma källa).

Vad som medvetet INTE portades:
- Agentens gamla Nuxt-gränssnitt (checkpoints, kanban, readme, roadmap, repo-map). Plan har egna sidor för mål, chatt och körningar. Koden finns kvar i agents-repot tills det arkiveras.
- Det gamla flödet med filbaserade godkännanden och Slack.

Lärdomar:
1. Ett kustomize-overlay kan inte referera en katalog utanför sig (LoadRestrictionsNone gäller resurser, inte kustomization-filer). Agentens manifesten lades därför i base, inte i dev - dev och prod delar nu samma källa.
2. En delete-patch i kustomize måste ligga på samma "nivå" ($patch: delete först i dokumentet, metadata med namespace) - annars träffar den ingenting och bygget fallerar med "no resource matches strategic merge patch".
3. När filer flyttas mellan repon ska symlänkar och byggverktyg följa med. wikictl hittar wikin genom att söka uppåt från arbetskatalogen, så det räckte att bygga om verktyget från det nya repot.
4. Ett repo med två deployables behöver en build per deployable, men en CI-workflow per modul - annars testar bara den ena.
