---
title: Delivery-pipeline för Plan: från dev till prod är nu dokumenterad
type: note
tags: [ci, deploy, pipeline, flux, dokumentation, projekt:plan]
created: 2026-11-27
updated: 2026-11-27
status: active
related: [notes/plan-repots-ci-och-prod-pipeline-var-trasig-fixade-sokvagar-och-bygg]
---

Plan-repots hela väg från dev till produktion är nu dokumenterad i
`docs/delivery-pipeline.md` (2026-11-27). Den samlar kedjan: lokal dev
(kind + Tilt + Neon) -> CI-grindar -> imagebygge till GHCR -> bot-PR som
byter image-raden i `environments/prod/patches/*.yaml` -> merge till main ->
Flux (GitRepository 1 min / Kustomization 5 min, path=environments/prod) ->
k8s-rollout. Grunden är oförändrad: **ingenting deployas direkt från en
pipeline** - pipelines bygger och committar, Flux applicerar. Därför är
produktionsläget alltid läsbart ur git, och rollback är en revert.

Varför: kedjan var utspridd över tre repon (plan, infrastructure-flux,
deployment-pipeline) och ingen enskild dokumentation beskrev hela vägen.
Den gamla wikinoten om "CI- och prod-pipeline var trasig" fångade bara
incidenten, inte normalflödet.

Viktigast, och nytt: **två luckor hittades när kedjan kartlades**, och båda
är verkliga risker i drift:

1. PR-preview deployar ingenting. `pr-preview.yaml` skapar overlays i
   `environments/pr/`, men ingen Flux Kustomization pekar på den sökvägen
   och klustret har inga `plan-pr-*`-namespaces. En PR-kommentar lovar en
   URL som inte finns.
2. Go- och Rust-flöden konkurrerar om samma rad i
   `environments/prod/patches/api-deployment.yaml` (båda använder `sed` på
   image-raden). Den senaste mergen vinner - en Go-build kan avinstallera
   Rust-imagen i prod. Detta är extra snällt just nu, eftersom dev kör Go
   och prod kör Rust.

Andra fynd: `.github/pr.yaml` och `.github/workflows/acceptance.yaml` är
döda (första ligger fel - GitHub läser bara `.github/workflows/`; den andra
refererar en katalog `frontend/` som inte finns). `environments/kubernetes/
overlays/dev` är en död Flux-overlay sedan dev gick över till Tilt.

Lärdom: när man kartlägger en leveranskedja är de intressanta fynden inte
i normalflödet utan i var *två* flöden skriver till samma ställe. Här
skrev Go- och Rust-pipelines till samma manifestrad - något ingen kunde se
i logiken utan att jämföra workflows sida vid sida.

## Relaterade

- [[notes/plan-repots-ci-och-prod-pipeline-var-trasig-fixade-sokvagar-och-bygg]]

