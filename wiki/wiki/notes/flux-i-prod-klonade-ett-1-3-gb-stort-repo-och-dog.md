---
title: Flux i prod klonade ett 1,3 GB stort repo och dog
type: note
summary: Flux GoGit-repoets stora PXE-träd fick source-controller att OOM:a; FluxInstance använder nu sparseCheckout för att bara hämta infrastructure-flux/.
tags: [flux, incident, prod, projekt:infrastructure]
created: 2026-11-24
updated: 2026-11-26
status: active
related: [concepts/flux-operator, entities/flux-gitops]
---

2026-11-24–26: Flux-källan för infrastructure-repot fastnade under uppdatering och source-controller blev OOM-killed. Den genererade Git-artifacten var 547 812 571 byte; nästan hela storleken kom från PXE-bootfiler. Kustomize-källan behöver bara infrastructure-flux/.

De tillfälliga resurs- och probe-override som noterades 2026-11-24 fanns inte i den spårade FluxInstance-konfigurationen den 26 november: live source-controller hade 1 GiB minnesgräns och GitRepository hade 60 s timeout. En manuell ändring av den genererade GitRepositoryn räcker inte heller, eftersom FluxOperator äger och återställer den.

Varaktig åtgärd 2026-11-26: infrastructure-flux/components/flux/flux-instance.yaml patchar den genererade GitRepositoryn med spec.sparseCheckout: [infrastructure-flux]. Samma begränsning finns som .sourceignore-fallback för artifact-filtering. Ändringen verifierades med flux-operator build instance och server-side dry-run. Efter den manuella appliceringen av samma FluxInstance-spec blev GitRepository Ready=True med observedSparseCheckout=[infrastructure-flux]; artifacten sjönk till 57 028 byte. Source-controller startade om tre gånger medan den gamla fullständiga källan byttes ut. Kustomize-controller startades om från worker-4 till worker-7; bootstrap, gateway och dot-ai synkade därefter den nya revisionen.

Kvarvarande fel är separata från källstorleken: root-steget stoppas av timeout mot Longhorn admission-webhook och resources-steget av timeout mot Kyverno-webhook. De kräver den nodspecifika nätverksfelsökningen, inte en ändring av sparse-checkout-fixen.

Relaterat: [[concepts/flux-operator]] och [[entities/flux-gitops]].
