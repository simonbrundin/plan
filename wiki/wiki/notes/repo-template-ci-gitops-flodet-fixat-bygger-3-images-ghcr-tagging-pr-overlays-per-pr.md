---
title: repo-template: CI/GitOps-flödet fixat (bygger 3 images, GHCR-tagging, PR-overlays per PR)
type: note
tags: [ci, github-actions, gitops, kustomize, flux, repo-template, bug, feature, projekt:repo-template]
created: 2026-11-03
updated: 2026-11-03
status: active
---

2026-11-03: Fixade det trasiga CI/GitOps-flödet i repo-template.

Var trasigt:
- ci-for-js-app.yaml pekade pa en gammal frontend/-katalog (fanns inte) och anropade en delad workflow med package-json-path: frontend.
- pr-preview.yaml byggde frontend/Dockerfile och genererade en overlay som patchade Deployment 'app' - men base har go-api/nuxt-app/rust-api, sa patchen traffade ingenting. Dessutom pekade HTTPRoute pa Service 'app' (fanns inte; servicen heter nuxt-app-service).
- environments/prod hade ingen image-override (alla images pekade pa :latest) och skapade TVa Namespaces ('repo' fran base + 'repo-template'), och apparna hamnade i default-ns.
- environments/pr/kustomization.yaml inkluderade base OCH per-PR-overlays som sjavla inkluderade ../../base -> dubbletter i kustomize.

Losning:
- Ny ci.yaml: bygger och pushar nuxt-app, go-api, rust-api till GHCR (taggar :<sha> och :main), och skriver sedan image-taggarna in i environments/prod/kustomization.yaml och committar - Flux deployar sedan.
- Ny pr-preview.yaml: bygger samma tre images med taggen pr-<NUM>-<sha> och genererar environments/pr/<NUM>/{kustomization.yaml,httproute.yaml} (base + images-transformer + namespace pr-<NUM> + egen HTTPRoute mot nuxt-app-service).
- environments/pr/kustomization.yaml innehaller nu bara per-PR-overlays (+ preview-info.yaml som platshallare sa att Flux-kustomizationen aldrig blir tom).
- environments/prod: namespace: repo-template, images-transformer mot GHCR, Namespace-etiketter via patch, och environments/prod/namespace.yaml + environments/pr/httproute.yaml ar borttagna (dubbletter).

Fallor:
- GitHub Actions YAML: en heredoc-rad pa kolumn 0 (t.ex. multi-rad commit-meddelande eller markdown-tabell i ett run: | -block) bryter YAML-parsningen - allt inuti run: | maste ha samma indenting. Losning: skapa filer med yaml.dump i python istallet for att bada in YAML-text.
- Kustomize images-transform: basens image-falt maste heta 'go-api:latest' (utan registry) for att Tilt ska kunna substituera mot lokal registry; GHCR-namnet satts bara i overlayna.
- Namespace-transformern DOPER om Namespace-objekt (base 'repo' -> 'repo-template' i prod), darfor kan inte bade base och overlay ha en egen Namespace med samma namn - tag bort overlayns Namespace och satt etiketter via patch istallet.
- TvA PR-overlays i samma parent-kustomization fungerar (olika namespace -> olika resource-ID:n), men varje overlay maste inkludera ../../base och parent fAr INTE inkludera base.

Verifierat lokalt: kustomize build ok for base/dev/prod/pr, och jag simulerade bade PR-overlay-generatorn och prod-tag-uppdateringen (python-blocket ur workflow-filen) - resultatet renderar korrekt med ratt image och namespace.
