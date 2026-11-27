---
title: Prod-patch utan namespace bryter Flux-builden
type: source
added: 2026-11-27
origin: https://github.com/simonbrundin/plan/pull/142
immutable: true
---

# Prod-patch utan namespace bryter Flux-builden

En strategisk merge-patch i `environments/prod/patches/` måste ange
`namespace` i `metadata` om resursen i basen har ett namespace.

Symtom: Flux Kustomization "plan" i produktion stannade med
`kustomize build failed: no resource matches strategic merge patch
"Deployment.v1.apps/agent-api.[noNs]"` - och inga nya bilder deployades,
eftersom builden aldrig lyckades.

Orsak: `environments/base/agent-deployment.yaml` sätter `namespace: plan`
på agent-API:t, men patchen i prod saknade namespace. Kustomize matchar
strategiska patchar på GVK + namn + namespace, så `[noNs]` matchade inget.

Fix: lägg till `namespace: plan` i patchens metadata.

Samma sak gäller andra base-resurser: `environments/base/api-deployment.yaml`
och `environments/base/deployment.yaml` har INTE namespace i sin metadata
(där sätts det senare), och deras prod-patchar fungerar därför utan.

Lärdom: när en patch slutar matcha är felet "ingen patchad resurs", inte
"fel bild". Verifiera alltid med `kustomize build environments/prod` lokalt
innan push - Flux ger ingen varning om det bara är en build.
