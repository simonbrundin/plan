---
title: API-servern i cluster1 spatar inte managedFields - SSA-konflikter upptacks inte
type: note
tags: [kubernetes, ssa, flux, managedfields, api-server, projekt:infrastructure]
---

I cluster1 spatar API-servern inte managedFields alls. Verifierat 2026-11-27 med en farsk server-side apply av en test-ConfigMap: applyen lyckades men metadata.managedFields var null pa det nyskapade objektet. Ingen av apiserverns argument innehaller --enable-managed-fields, sa det ar serverns byggda standard i Kubernetes v1.37.0.

Varfor det spelar roll: server-side apply-konflikter upptacks inte. Flux kan da tyst overgora fält som nagons annan aktor agg, och "managedFields sag vem som ager vad" ar inte en giltig felsokningsmetod i detta kluster. Tidigare sessioner i detta kluster hade faktiskt sett managedFields med talos/Apply och kubectl-client-side-apply pa CoreDNS-ConfigMapen, sa sparningen har forsvunnit sedan dess - troligen nar apiservrarna startades om 2026-11-25.

Konsekvens for arbetssattet: agar måste uttryckligen designas, inte furas pa SSA-konfliktdetektering. En aktor per objekt (Talos ager CoreDNS-Deploymenten, Flux ager CoreDNS-ConfigMapens Corefile, Longhorns manager ager standarddiskens storageReserved, Flux ager resten av Longhorn Node-CR:en).

Verifiera alltid med en faktisk server-side dry-run, inte med managedFields.
