---
title: Worker-3 PoE-omstart och Talos-konfiguration 2026-11-20
type: note
tags: [talos, worker-3, poe, felsokning, projekt:infrastructure]
---

Worker-3 på 10.10.10.23 startades om via UniFi US48P500 port 23. Hela port_overrides-listan bevarades vid PUT; porten cyklades off i 12 sekunder och tillbaka till auto. Noden svarar nu på Talos API och accepterar klusterkonfigurationen från pxe/matchbox/assets/talos-config/nodes/10.10.10.23.yaml. En DHCP-konfiguration för end0 lades till och noden power-cyklades igen. Talos tjänster machined/containerd/cri är healthy, men kubelet startar inte: exec /usr/local/bin/kubelet: input/output error. Därför syns worker-3 ännu inte i Kubernetes. Volymerna på /dev/mmcblk1 är monterade och inga tydliga mmc I/O-fel syntes i dmesg; nästa steg är att felsöka/reflasha microSD-kortet eller verifiera kortets läsbarhet.
