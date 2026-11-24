---
title: worker-1 svarar inte ens med en full PoE-cykel - noll paket i bade riktningar
type: note
summary: Status 2026-10-27: en verifierad PoE-cykel pa UniFi port 21 (45 s av, 45 s up) gav PHY-link men fortfarande noll rx-paket, tom MAC-tabell och noll DHCP/TFTP...
tags: [pxe, hardware, felsokning, rock4cplus, unifi, projekt:infrastructure]
created: 2026-10-27
updated: 2026-10-27
status: active
related: [entities/pxe-server, notes/worker-1-och-worker-2-ar-doda-pa-link-niva-pxe-kedjan-ar-inte-blockerande]
---

Status 2026-10-27: en fullt verifierad PoE-cykel pa UniFi port 21 (worker-1,
`10.10.10.21`, MAC `0a:3f:39:25:2b:c8`) gav PHY-link men **fortfarande noll
trafik fran noden**.

## Vad testet visade

- `poe_mode=off` -> `up=false`, `pwr=0.00 W`, `link_down_count` 4 -> 5.
- `poe_mode=auto` -> `up=true`, `speed=1000`, `pwr=1.84 W`.
- `rx_packets=0` hela tiden, `mac_table_count=0` (switchen larde aldrig in
  nodens MAC pa porten).
- Tcpdump pa controlplane-1 (samma VLAN 2, iface `enabcm6e4ei0`) under hela
  cykeln: **114 paket, alla ARP-fragor fra gatewayn** (`0e:ea:14:56:90:e6`,
  "who-has 10.10.10.21"). **Noll paket fran `0a:3f:39:25:2b:c8`**, noll DHCP,
  noll TFTP, noll HTTP mot matchbox.

## Varfor detta bevisar att felet inte ligger i PXE-kedjan

Tre oberoende observationer pekar samma hall:

1. **Switchen laddar broadcast till port 21** i exakt samma omfattning som
   till den fungerande porten 23 (2.93 M mot 2.94 M broadcast-paket). VLAN 2
   nar alltsa fram till porten - porten ar inte isolerad och har ratt VLAN.
2. **Gatewayn fragar aktivt ARP for `10.10.10.21`** i rak strum, men noden
   svarar aldrig.
3. **Stramforbrukningen avslajar**: port 21 drar 1.84 W, port 22 (worker-2)
   1.99 W, medan levande workers drar 3.6-4.9 W. Boarden far stram men
   startar inte en nativ gang - den ligger i PHY-lank utan att SoC:en
   agerar.

Port 21 har dessutom en enda `port_overrides`-rad med
`native_networkconf_id=676b3948f5c488156c0d3c99` ("Default", 10.1.1.1/24)
medan alla overiga worker-portar har `native=null` (arv). Detta foljde med
den override-rad som skapades av PoE-styrningen och ar **inte** orsaken:
broadcastflodet visar att porten ligger i samma segment som port 23.

## Varfor lopet inte kan observeras via kubectl exec

`kubectl exec` mot pods pa controlplane-1/2/3 ger `Unable to connect to the
server: EOF`, medan `kubectl get` och `kubectl logs` fungerar. Orsaken ar att
apiservern streamar via kubelet pa port 10250, och just dessa noder har
kubeletcertifikat signerade av en **gammal Kubernetes-CA** (se
[[notes/worker-7-s-kubeletcertifikat-signerades-av-en-gammal-kubernetes-ca-som-finns-kvar-i-git-historiken]]).
Felsokning maste ga via Talos API:t (`talosctl read`) i stallet for
`kubectl exec` - fungerar stabilt.

## Slutsats

worker-1 (och med samma tecken worker-2) ar **inte** ett PXE-, DHCP- eller
VLAN-problem. Boarden far stram och forhandlar Ethernet men skickar inga
egna paket. Nasta steg ar fysisk atkomst: U-Boot-konsol via serie for att
verifiera `boot_targets`/`boot_script_dhcp`, eller microSD-kortstestet
samma kort i worker-3 (se
[[notes/worker-1-och-worker-2-ar-doda-pa-link-niva-pxe-kedjan-ar-inte-blockerande]]).

## Relaterade

- [[entities/pxe-server]]
- [[notes/worker-1-och-worker-2-ar-doda-pa-link-niva-pxe-kedjan-ar-inte-blockerande]]
- [[notes/worker-7-s-kubeletcertifikat-signerades-av-en-gammal-kubernetes-ca-som-finns-kvar-i-git-historiken]]
