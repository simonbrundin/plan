---
title: Talos-varningar pa controlplanes: 90% kommer fra Talos Discovery Service + DNS
type: note
tags: [talos, felsokning, dns, discovery, projekt:infrastructure]
---

Diagnos 2026-11-20 (session: "kolla talos-loggarna for controlplanes, finns det nagot att fixa?").

## 1. Var varningarna kommer ifran (cp1: 1265 warnings i loggen)

| Kategori | Antal | Rotorsak |
|---|---|---|
| "error serving dns request" (udp -> 10.10.10.1:53, i/o timeout) | 770 | 767 av 770 (99.5%) kommer inom 15 s efter en discovery-rad -> DNS-fel pa grund av Talos Discovery Service |
| "error watching discovery service state" / "hello failed ... discovery.talos.dev" | 402 | discovery.talos.dev (18.226.100.232:443) gar inte att naa fran noderna |
| "network is unreachable" mot 8.8.8.8 / 1.1.1.1 | 56 | Talos default-resolvrar anvands innan DHCP-leasen (ingen defaultroute) - bara vid boot |
| NTS-tidsynkfel (IPv6 "network is unreachable") | ~10 | noderna har global IPv6 men ingen IPv6-defaultroute; NTS forsoker IPv6 |
| link/address/nodeIP-fel ("no suitable node IP found", "removed address", endpoint watch) | ~30 | link-flap: alla tre controlplanes tapte LAN 2026-11-20 12:06:25 i ~3 s (UniFi-switch, port 11-13) och 2026-11-18 02:04 (switchen rebootade, uptime 220979 s = boot ~2026-11-18 03:00) |

Verifierat med python-korrelation: "dns timeouts within 15s of discovery: 767/770", burst: 56/56 bursts
nara en discovery-rad.

## 2. Viktigt: UDP fra Talos-noderna ar opalitligt - problemet ar nod-lokalt

- Samma LAN, samma gateway (10.10.10.1) och samma mal: PXE-varden 10.10.10.2
  (ssh pxe) klarar NTP (udp/123) 10/10 och DNS mot 10.10.10.1 pa <50 ms,
  samt HTTPS till discovery.talos.dev (200, TLS 0.6 s).
- Talos-controlplanet timeout:ar samma UDP-floden (udp/123 och udp/53).
  TCP fungerar (etcd, apiserver, HTTPS).
- Alltsa: inte router-/brandvaggproblem - det ligger i nodens stack/NIC
  (bcmgenet pa RPi4) eller i Talos resolver. INTE utredd - noterad som
  "inte-obvious bugg" for framtiden.

## 3. Fixar applicerade 2026-11-20 (pa alla tre controlplanes, --mode no-reboot)

- ResolverConfig: nameservers=[10.10.10.1] (explicit LAN-resolver i stallet
  for Talos default 1.1.1.1/8.8.8.8).
- TimeSyncConfig: useNTS=false, NTP pa IP (162.159.200.1, 216.239.35.0),
  bootTimeout 5m - tidsynk utan DNS och utan IPv6.

Resultat: 0 nya warnings pa alla tre noderna efter patchen; noderna Ready,
etcd 3/3 i synk, inga k8s Warning-events.

## 4. Metodfakta (viktiga)

- `talosctl patch machineconfig` SLAR IHOP listor (append). Att patcha
  nameservers två ganger ger dubbla entries. Losningen: patcha med
  "$patch: delete" pa dokumentet och sedan samma dokument med nytt innehall
  (funkar, och da blir patchen idempotent).
- I Talos v1.14 ligger nameservers INTE i machine.network (avvisas med
  "already set in v1alpha1 config") - de ligger i ResolverConfig-dokumentet.
- machine.features.discoveryService finns INTE i v1.14. Discovery stangs av
  genom att TA BORT alla DiscoveryServiceConfig-dokument (patch: "$patch: delete").
  KubePrism (port 7445, aktiverad i all.yaml) kräver discovery - att stanga
  av den är en avvägning, inte gjort.
- `talosctl patch machineconfig --dry-run` fungerar for att validera en patch
  utan att applicera den.
- Tiden: node timestamps i kubelet-loggen ar epoch-ms ("ts":1795176386378).

## 5. Kvarstående rekommendationer (inte gjorda)

1. Decidera om discovery-tjansten ska bort (>$95% av varningarna). Da maste
   KubePrism-bedomningen goras samtidigt.
2. `talos/patches/all.yaml` har ALDRIG applicerats pa controlplanet (live-config
   fran 2026-08-31 saknar certSANs och rotate-server-certificates).
   Applicera all.yaml i sin helhet vid ett tillfalle (kraver apiserver-
   cert-rotation -> controlplane-pod-restart).
3. UDP-felet pa noderna bor utredas (bcmgenet/checksum-offload?) - det star
   med en fungerande worker.
4. Switchen flappade 2026-11-20 12:06 pa port 11-13 (alla controlplanes
   samtidigt) - undersok switchens PoE/upscalning.
