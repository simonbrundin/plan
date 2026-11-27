---
title: Prod-patch utan namespace bryter Flux-builden
type: note
tags: [bug, flux, kustomize, projekt:plan]
---

En strategisk merge-patch i environments/prod/patches/ måste ange namespace i metadata om resursen i basen har ett namespace - annars matchar kustomize ingenting.

Symtom i produktion: Flux Kustomization 'plan' fastnade med 'no resource matches strategic merge patch Deployment.v1.apps/agent-api.[noNs]', så inga nya bilder deployades över huvudet på (t.ex. den Nuxt-bild som hade fixen för tangenten s).

Orsak: environments/base/agent-deployment.yaml satte namespace: plan, men patchen saknade namespace. Fix: lägg till namespace: plan i patchens metadata (PR #142).

Lärdom: kör 'kustomize build environments/prod' lokalt innan push - en trasig patch ger bara 'ingen patchad resurs' i Flux-loggen, utan att peka på den felaktiga filen.
