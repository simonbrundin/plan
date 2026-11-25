---
title: Klustret har byggts om - bara kube-system, Flux och appar saknas (2026-11-23)
type: note
tags: [kubernetes, talos, gitops, projekt:infrastructure]
created: 2026-11-23
updated: 2026-11-25
status: active
related: [notes/prod-klustret-saknar-quorum-controlplane-1-och-2-svarar-inte]
---

Observation 2026-11-23 kl 22:00: klustret (context admin@cluster1-2) har ENDAST kube-system. Namespaces: default, kube-node-lease, kube-public, kube-system. 'kubectl get crd' returnerar noll CRDs - varken FluxOperator, cert-manager, Longhorn, Envoy Gateway, CNPG, Zitadel eller ESO finns. Bara kubeadm-komponenterna (apiserver, scheduler, controller-manager, flannel, coredns, kube-proxy) ror sig, och pa controlplane-1 star de statiska podarna Terminating (apiserver/controller-manager/scheduler) medan cp-2 och cp-3 startade for ca 4-5 timmar sedan.

Varfor det spelar roll: wikin (entities/kubernetes) beskriver ett fullt utrustat kluster med Flux, Longhorn, Envoy Gateway, Prometheus och alla appar. Det beskriver arkitekturen i infrastructure-flux/ och ar korrekt for repot - men inte klustrets aktuella tillstand. En agent som felsoker 'varfor svarar inte plan.simonbrundin.com' eller 'var synkar Flux inte' kan inte skylla pa appfel; forst maste FluxOperator bootstrapas in i det nya klustret.

Anteckning: noder controlplane-1..3 + worker-3..8 ar Ready (worker-1, worker-2, worker-9 saknas). Minnesanteckningen 'cluster1-stale-kubeconfig-ca-och-noll-workers-diagnos-2026-11-11' galler en tidigare diagnos, inte detta.

Nasta steg (att bekrafta med anvandaren): bootstrap av FluxOperator (flux-operator + FluxInstance ur infrastructure-flux/components/flux/flux-instance.yaml) och sedan synk av infrastructure-flux. Tills dess ar det 'Flux kan inte synka' inte en bugg utan en saknad installation.

## Relaterade

- [[notes/prod-klustret-saknar-quorum-controlplane-1-och-2-svarar-inte]]

