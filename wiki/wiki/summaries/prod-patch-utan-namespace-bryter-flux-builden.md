---
title: Sammanfattning: prod-patch utan namespace bryter Flux-builden
type: summary
summary: En strategisk merge-patch i environments/prod/patches/ måste ange namespace i metadata om resursen i basen har ett namespace - annars matchar kustomize ingenting och Flux stannar med en misslyckad build.
tags: [flux, kustomize, bug, projekt:plan]
sources: [raw/2026-11-27-prod-patch-utan-namespace-bryter-flux-builden.md]
created: 2026-11-27
updated: 2026-11-27
status: active
related: [notes/prod-patch-utan-namespace-bryter-flux-builden]
---

## Källa

`raw/2026-11-27-prod-patch-utan-namespace-bryter-flux-builden.md` (origin: https://github.com/simonbrundin/plan/pull/142)

## Huvudpunkter

1. **Symtomet är "ingen patchad resurs", inte "fel bild".** Flux
   Kustomization `plan` i produktion fastnade med `no resource matches
   strategic merge patch "Deployment.v1.apps/agent-api.[noNs]"`. Eftersom
   builden aldrig lyckades deployades inga nya bilder alls - även om
   GitHub Actions publicerade och öppnade PR:er med nya bilder.
2. **Orsak: saknad `namespace` i patchen.** `environments/base/agent-deployment.yaml`
   sätter `namespace: plan` på agent-API:t, men patchen i
   `environments/prod/patches/agent-deployment.yaml` saknade namespace.
   Kustomize matchar strategiska patchar på GVK + namn + namespace, så
   `[noNs]` matchade ingenting - och felet pekade inte ut den felaktiga filen.
3. **Fix:** lägg till `namespace: plan` i patchens `metadata`.
4. **Samma sak gäller andra base-resurser:** `environments/base/api-deployment.yaml`
   och `environments/base/deployment.yaml` har INTE namespace i sin metadata
   (det sätts senare), och deras prod-patchar fungerar därför utan.
5. **Verifikationsregel:** kör `kustomize build environments/prod` lokalt
   innan push. En trasig patch ger bara en byggmisslyckning i Flux-loggen,
   utan hjälp om var filen ligger.

## Kopplingar

- [[notes/prod-patch-utan-namespace-bryter-flux-builden]]
