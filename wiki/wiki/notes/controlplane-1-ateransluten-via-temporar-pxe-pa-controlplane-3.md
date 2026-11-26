---
title: controlplane-1 återansluten via temporär PXE på controlplane-3
type: note
tags: [incident, talos, pxe, recovery, projekt:infrastructure]
created: 2026-11-26
updated: 2026-11-26
status: active
---

controlplane-1 (10.10.10.11) fastnade i Raspberry Pi DHCP/TFTP-boot när PXE-värden 10.10.10.2 var död. En CP1-begränsad dnsmasq/TFTP-responder kördes tillfälligt i hostNetwork-podden sniff3 på controlplane-3 och serverade verifierade filer från Raspberry Pi-backupen samt matchbox. Efter PoE-cykel på switchport 11 fick CP1 DHCPACK, startade Talos och Kubernetes. CP1 blev Ready, Talos API svarar och den befintliga etcd-medlemmen 50b0f7c3297c1e4e återkom; ingen etcd-medlem togs bort. etcd har quorum med leader 199a9f61a6dc6e1b. De temporära sniff-poddarna togs bort efteråt. UniFi-nyckeln som låg i /tmp/unifi.key raderades; den exponerade nyckeln måste fortfarande roteras.
