---
title: DNS och DHCP har skilda agar: dnsmasq pa pxe ager DHCP, UniFi pa 10.10.10.1 ager DNS - och worker-9:s lease-hostname ar klientens eget
type: note
tags: [dns, dhcp, unifi, dnsmasq, worker-9, coredns, projekt:infrastructure]
created: 2026-11-27
updated: 2026-11-27
status: active
related: [entities/pxe-server, entities/talos-linux, notes/worker-2-och-worker-9-saknas-fortfarande-som-cluster1-workers-diagnos-2026-11-26]
---

Kartlagt 2026-11-27.

Agarskap: DHCP utlosas av dnsmasq (containern dnsmasq-proxy, image dockurr/dnsmasq, pa pxe = 10.10.10.2, livefil /home/simon/PXE/dnsmasq.conf, range 10.10.10.50-200). DNS utlosas av UniFi pa 10.10.10.1. Det ar TVa olika tjanster - dnsmasq har port=0 och kan inte leverera DNS, och klienterna pekas pa 10.10.10.1 som resolver.

UniFi saknar A-poster for worker-1 (10.10.10.21), worker-2 (10.10.10.22) och worker-9 (10.10.10.29), men har controlplane-1..3 och worker-3..8. Luckan ar strukturell, inte en pod eller en cache.

worker-9:s DHCP-lease har hostname MinisomhorSIMON. Det ar INTE ett dnsmasq-fel: dnsmasq anvander klientens eget hostname for lease, och maskinen kor fortfarande macOS, inte Talos. set:worker-9 i dnsmasq.conf ar en DHCP-tagg och styr inte leasens hostname. Nar maskinen bootar Talos skickar Talos sitt konfigurerade hostname (worker-9) i DHCP-requesten - samma sak som for worker-1..8, vilkas leases har riktiga hostname.

Varfor detta brast noderna: kubelet-serving-cert-approvern har bypassDnsResolution: false och slar upp nodens DNS-namn via klustrets DNS, alltsa CoreDNS - inte via UniFi. UniFi-luckan traffade altsa aldrig approvern; det var CoreDNS hosts-blocken som saknades. approverns loggar visar Denying

## Relaterade

- [[entities/talos-linux]]
- [[notes/worker-2-och-worker-9-saknas-fortfarande-som-cluster1-workers-diagnos-2026-11-26]]
- [[entities/pxe-server]]
