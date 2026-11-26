---
title: Tillfälliga felsökningspoddar rensade
type: note
tags: [cleanup, cluster1, debugging, projekt:infrastructure]
created: 2026-11-26
updated: 2026-11-26
status: active
---

Inventering av cluster1 hittade kvarvarande unowned debug-poddar i default, flux-system och kube-system: np-controlplane-2/3, np-worker-3, nt4, nt6, cloneprobe, sshgo, ssh-probe*, flannel-probe, netprobe*, nettest*, nt-worker-4. De körde endast tillfälliga ping/TCP/DNS/SSH/Git/Flannel-tester eller sleep och togs bort. Efter rensningen finns inga unowned pods kvar. Controller-ägda workload-pods, historiska Job-resurser och Longhorn/Flux-resurser lämnades orörda.
