---
title: worker-4 återansluten till cluster1 via lokalt SD-kort
type: note
tags: [talos, worker-4, incident, återställning, projekt:infrastructure]
---

worker-4 bootade från sitt lokala SD-kort utan PXE. Kortets gamla generation-1 STATE/config.yaml togs bort efter backup på PXE-servern. Efter PoE-cykel applicerades den aktuella worker-4-konfigurationen via Talos maintenance API med talosctl apply-config --insecure. Noden anslöt till cluster1 som Talos v1.14.1 och blev Kubernetes Ready; kube-flannel och kube-proxy blev Running. Viktig felsökningsdetalj: talosctl-flaggor måste placeras efter kommandot, exempelvis talosctl get members -n 10.10.10.24 --insecure; den tidigare vakaren använde fel flaggordning.
