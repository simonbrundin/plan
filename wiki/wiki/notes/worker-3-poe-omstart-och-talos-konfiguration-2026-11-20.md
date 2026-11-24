---
title: Worker-3 PoE-omstart och Talos-konfiguration 2026-11-20
type: note
tags: [felsokning, poe, projekt:agents, projekt:infrastructure, talos, worker-3]
created: 2026-11-20
updated: 2026-11-20
status: active
---

Worker-3 på 10.10.10.23 startades om via UniFi US48P500 port 23. Hela port_overrides-listan bevarades vid PUT; porten cyklades off i 12 sekunder och tillbaka till auto. Noden svarar nu på Talos API och accepterar klusterkonfigurationen från pxe/matchbox/assets/talos-config/nodes/10.10.10.23.yaml. En DHCP-konfiguration för end0 lades till och noden power-cyklades igen. Talos tjänster machined/containerd/cri är healthy, men kubelet startar inte: exec /usr/local/bin/kubelet: input/output error. Därför syns worker-3 ännu inte i Kubernetes. Volymerna på /dev/mmcblk1 är monterade och inga tydliga mmc I/O-fel syntes i dmesg; nästa steg är att felsöka/reflasha microSD-kortet eller verifiera kortets läsbarhet.

## Uppdatering 2026-11-20

Worker-3 på 10.10.10.23 startades om via UniFi US48P500 port 23. Hela port_overrides-listan bevarades vid PUT; porten cyklades off i 12 sekunder och tillbaka till auto. Noden svarar på Talos API och accepterade klusterkonfigurationen från pxe/matchbox/assets/talos-config/nodes/10.10.10.23.yaml. En DHCP-konfiguration för end0 lades till och noden power-cyklades igen.

Symtom: Talos-tjänsterna machined/containerd/cri var healthy, men kubelet startade om i en loop med 'exec /usr/local/bin/kubelet: input/output error' och noden syntes inte i Kubernetes.

Felsökningen utesluter bland annat artefakt- och arkitekturfel (kubelet-image ghcr.io/siderolabs/kubelet:v1.37.0, binärens SHA-256 7855ce2c3a3393d3c871fdeedcef32e38531f6852179130764763a9da79e2aa8 matchar controlplane-1), nätverk/DHCP, Talos-API, kluster-CA och att alla containerd-content-blobs matchade sina digester. En minimal alpine-container startade felfritt i både inmem- och system-namespace, vilket utesluter generell containerd-/overlayfs-fel.

Lösning: 'talosctl reset --nodes 10.10.10.23 --graceful=false --reboot=true --system-labels-to-wipe=EPHEMERAL' tömde /var (XFS på /dev/mmcblk1p6 på microSD-kortet) och byggde om containertilståndet. Efter reboot var kubelet healthy och worker-3 kom in som Ready i Kubernetes (v1.37.0, Talos v1.14.1, arm64).

Slutsats: EIO-felet kom från korrupt containerd-state/XFS på microSD-kortet, inte från artefakten eller runtime-konfigurationen. Rätt verktyg för den typen av fel är att tömma EPHEMERAL, inte att reflascha installationen.

Kvarvarande separat fråga: volymen u-ssd-1 på /dev/sda fallerar med 'filesystem type mismatch: gpt != xfs' eftersom disken har en GPT-partitionstabell men Talos förväntar sig XFS.

