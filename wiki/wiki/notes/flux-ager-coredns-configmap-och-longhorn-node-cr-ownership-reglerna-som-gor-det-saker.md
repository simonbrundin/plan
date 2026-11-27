---
title: Flux ager CoreDNS-ConfigMap och Longhorn Node-CR - ownership-reglerna som gor det saker
type: note
tags: [flux, coredns, longhorn, dns, ssa, projekt:infrastructure]
created: 2026-11-27
updated: 2026-11-27
status: active
related: [concepts/flux-operator, concepts/longhorn, summaries/infrastructure-flux]
---

Kommittat 2026-11-27: 6c7c84f9 (talos) + 6551803b (flux).

Varfor: TvA resurser hade bara livsversion. CoreDNS Corefile var patchad manuellt for att UniFi saknar nodposter for worker-1/2/9, och worker-1:s Longhorn Node var registrerad manuellt via SSA - en handpagaing eller Longhorn-upgrade kunde radera den.

Losningen: Flux ager nu bada, i infrastructure-flux/components/{coredns,longhorn-system}/resources/, synkade av resources-steget. Inget skyddat aggregator-fila (components/kustomization.yaml, apps/kustomization.yaml) behovdeandras.

De fyra icke-obvious reglerna som gor att det far sta:

1. Talos applicerar ALDRIG en ConfigMap som redan finns i klustret. Talos v1.14.6:s manifest_apply gor forst inventory-koll (skip) och sedan en Get - om resursen finns backfillas inventariet och steget hoppas over. Darrfor kan Flux aga .data.Corefile utan konflikt, och Talos ager fortfarande Deploymenten och annotationen config.k8s.io/owning-inventory (den lammas ute ur Flux-manifestet sa att Talos behaller agarskapet).

2. Longhorns spec.disks ar en GRANULAR map - CRD:en saknar x-kubernetes-map-type, och SSA merge:ar per nyckel. Bevis: en server-side dry-run av ett manifest med ENDAST ssd-1 bevarade default-disk-b30600000000, och managed fields delar agarskapet per nyckel (kustomize-controller: f:spec.f:disks.f:ssd-1, longhorn-manager: f:spec.f:disks.f:default-disk...). En atomisk map hade tagit bort standarddisken.

3. Longhorns validator avvisar en uppdatering om en befintlig disk FORSVINNER ur spec.disks sa lange den har allowScheduling true eller schemalagt utrymme (Please

## Relaterade

- [[summaries/infrastructure-flux]]
- [[concepts/flux-operator]]
- [[concepts/longhorn]]
