---
title: Cluster1: Longhorn worker-4 återställd och Flux stabiliserad
type: note
tags: [cluster1, flux, longhorn, incident, projekt:infrastructure]
created: 2026-11-26
updated: 2026-11-26
status: active
---

Worker-4 hade intermittenta CRI/kubelet-problem och Longhorn engine-image kunde inte starta. Kubelet/CRI återstartades via talosctl; worker-4 blev Ready igen, Longhorn manager och engine-image blev Running, CNPG/Zitadel-databasen stabiliserades. Source-controller flyttades deklarativt till controlplane-1 via FluxInstance patch och Kyverno admission-controller pinades deklarativt till controlplane-1 i komponentens HelmRelease. Dashboard-CNPG-volymen kunde attachas efter höjning av Longhorn storage-over-provisioning till 200. Dashboard-appens HTTP-prober ersattes i dashboard-repot av TCP-prober och pushades som 304018a. Homelab är fortfarande blockerat av arr-stack-media: Longhorn-volume 100Gi, tre replika, faulted, actualSize 0, inga replica-objekt och storage scheduling failure. Ändra inte volymens repliker/storlek eller radera PVC utan beslut om dataförlust/storageexpansion.
