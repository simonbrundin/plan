---
title: worker-8 åter Ready, worker-1 och worker-2 saknar fysisk länk
type: note
tags: [incident, talos, workers, unifi, projekt:infrastructure]
created: 2026-11-26
updated: 2026-11-26
status: active
---

Efter CP1-återställningen power-cyklades UniFi-portarna 21, 22 och 28 med hela port_overrides-listan bevarad. worker-8 (10.10.10.28) kom tillbaka: Talos API och kubelet svarar, Kubernetes visar Ready. Port 28 har fysisk länk på 1 Gbit/s. worker-1 och worker-2 kom inte tillbaka: port 21 och 22 är enabled men up=false, speed=0, rx_bytes=0 och poe_good=false trots PoE auto. Talos API på 10.10.10.21/.22 är otillgängligt. Detta kräver fysisk kontroll av ström, kabel och rätt switchport; ingen ytterligare fjärråtgärd bör göras innan länk syns.
