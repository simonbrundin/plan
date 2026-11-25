---
title: Agents-repot arkiverat och mergens mål kompletterade
type: note
tags: [arkiv, repo, agent, ci, migrering, projekt:simon]
created: 2026-11-25
updated: 2026-11-25
status: active
related: [notes/sammanslagning-av-agents-repot-in-i-plan-vad-som-flyttades-och-vad-som-medvetet-inte-gjorde-det]
---

2026-11-25: agents-repot är arkiverat och sammanslagningen med plan-repot är helt avslutad.

Vad som gjordes:
- Det som låg ocommittat i agents-repot (60 ändrade filer + ~70 nya, bland annat det gamla Nuxt-gränssnittet och wikinoter) lades på branchen `archive/pre-merge-leftovers` och pushades. Ingenting kastades bort. main återställdes sedan och fick en arkivbanner i README.md och AGENTS.md som pekar om till plan-repot.
- Båda repona pushades: agents main (2 commits) och plan main (4 commits).
- Worktree `/tmp/agents-pre` prunades.

En verklig lucka hittades och fixades i plan: agent-imagen byggdes med **context = repo-roten** och repot hade ingen `.dockerignore`. `src/nuxt/node_modules` (över 1 GB) skickades med i varje bygge. Agents-repot hade en `.dockerignore` som nu är portad och anpassad (`plan/.dockerignore`) - contexten är 1,33 MB i stället för över 1 GB. Verifierat med en lokal `docker build` av agent-imagen.

En bugg fixades samtidigt: `environments/dev/Tiltfile` angav `resource_deps=["go-api"]` men resursen heter `go` (efter `new_name`), så Tilt ignorerade beroendet och agent-API:t startade utan go-API:t. Nu är det `resource_deps=["go"]`.

Mål som markerades klara (status_updates skrevs i historiken):
- #486 Repoordning: agents och plan
- #487 Beslut: monorepo eller två repos
- #488 Subtree av agents/src/go i plan
- #489 Flux deployar agent-api från plan
- #521 Bevara appstrukturen
- #523 Gemensam dev-miljö
- #524 Per-app build och CI
- #525 Slutlig systemverifiering (kodmål #540 med godkänt godkännande och livsplaneringsmål #543 genom samma AgentRun/Job-flöde)

#497 (Bygg sammanhållen planerings- och agentplattform) står kvar i `verify` - det som återstår är proddriften (hemligheter via `environments/prod/secrets/create-agent-secrets.sh`, borttagning av patcharna i `environments/prod/patches`, och CI som publicerar agent-imagen). #522 (Gemensamma kontrakt mellan Plan och agenten) står kvar i `plan` - det arbetet är inte gjort.

Varför det här var värt att göra ordentligt: ett ocommittat och oarkiverat agents-repo är en fälla för framtida agenter - de ändrar kod som aldrig deployas och förstår inte att wikin flyttat. Att lämna en ren main med en tydlig arkivbanner plus en branch med det gamla innehållet gör läget entydigt utan att förlora historik.

Relaterat: [[notes/sammanslagning-av-agents-repot-in-i-plan-vad-som-flyttades-och-vad-som-medvetet-inte-gjorde-det]]
