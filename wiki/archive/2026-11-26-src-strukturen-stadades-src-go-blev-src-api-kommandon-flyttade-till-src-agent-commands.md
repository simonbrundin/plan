---
title: src-strukturen städades: src/go blev src/api, kommandon flyttade till src/agent/commands
type: note
tags: [struktur, refaktorering, agent, projekt:plan]
---

Varför: src/ bestod av sex mappar med oklar namn - src/go (Plan API), src/agent (agenten, ALSO Go), src/nuxt, src/commands, src/contracts och src/api som bara innehöll en trasig Dockerfile.dev. Namnet 'go' på en tjänst (inte språket) var den verkliga förvirringen.

Vad som gjordes (2026-11-26):
- src/go -> src/api (git mv). Plan API:t heter nu 'api' precis som agent-imagen och k8s-manifesten (api-deployment.yaml) gör. Uppdaterade: .github/workflows/commit.api.yaml (paths, working-directory, docker context), environments/dev/Tiltfile (docker_build), AGENTS.md, docs/development.md, docs/agent-platform.md, environments/dev/README.md, src/contracts/plan-agent.yaml, in-app docs i src/nuxt/app/content/docs/ samt openspec-förslag.
- src/commands -> src/agent/commands. Bara agentmotorn laddar kommandona (Dockerfile COPY + registry.go). registry.go:s DefaultCommandsPath() söker nu 'src/agent/commands' (med 'src/commands' som legacy-kandidat), registry_test.go uppdaterad, .dockerignore och commit.agent.yaml justerade.
- Döda filer borttagna: src/api/Dockerfile.dev (refererade ./cmd/server/main.go i en katalog utan go.mod), src/go/server (49 MB kompilerad binär incheckad i git), src/nuxt/package-lock.json (bun används, ingen referens).
- src/agent/goals.json: 10 stale sökvägar 'src/go/internal/core/...' -> 'src/agent/internal/core/...'.

Varför INTE slå ihop agenten och API:t i en Go-modul: de har separata go.mod (olika dependency-grafer), separata CI-gates (commit.api.yaml vs agent-go.yml/commit.agent.yaml), separata bilder och separata deploycykler (environments/base/api-* vs agent-*). Ett 'nuxt + go'-repo skulle kräva att agenten rör om plan-API:t vid varje agentrelease.

Verifierat: go build/vet/test i src/api (alla paket ok), go build/vet/test i src/agent (alla paket ok), bun test i src/nuxt (233 pass), bun run build, docker build av agent-imagen (COPY src/agent/commands -> /app/src/agent/commands), DefaultCommandsPath() returnerar /home/simon/repos/plan/src/agent/commands.

Följd: wikin uppdaterad med de nya sökvägarna (bara plan-repots egna; andra repos pappa/learn/mailmaster som också har src/go lämnades orörda). wiki/raw/ är immutabel och lämnades orörd.
