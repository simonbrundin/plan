---
title: Talos: hostname: auto i en generisk config byter nodens namn och kastar ut den ur klustret
type: note
tags: [talos, hostname, config, kluster, felsokning, projekt:infrastructure]
created: 2026-10-24
updated: 2026-10-24
status: active
---

Misstag: applicerade pxe/matchbox/assets/talos-config/controlplane.yaml pa
controlplane-3 (10.10.10.13). Filen har HostnameConfig med "auto: stable",
vilket INTE ar nodens namn - noden fick namnet "talos-myp-y75" och den gamla
Kubernetes-noden "controlplane-3" blev stale (NotReady).

Symtom: klustret visar DUA noder for samma maskin:
  controlplane-3   NotReady
  talos-myp-y75    Ready
Bara en av dem ar den verkliga maskinen (samma DHCP-reservation,
e4:5f:01:a0:6d:d2 = pi4-3 i pxe/dnsmasq.conf).

Fix: applicera en config med hostname: controlplane-3, starta om noden, och
radera det gamla Kubernetes-nodnamnet:
  kubectl delete node talos-myp-y75

LARDOM: en generisk config med "auto" hostname ar bara for oinstallerade
noder. Se alltid nodens hostname innan du applicerar en config over den:
  talosctl get hostname -n <ip>
