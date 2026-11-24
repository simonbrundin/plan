---
title: Twingate-anslutaren for 10.10.10.0/24 var nere - recovery-floden kan inte koras fran arbetsstationen
type: note
tags: [incident, pxe, worker-1, twingate, projekt:infrastructure]
---

Varfor: under forsoket att aterstalla worker-1 (10.10.10.21) visade det sig att hela datacentersegmentet 10.10.10.0/24 var onatbart fran arbetsstationen: TCP-anslutningar till 10.10.10.2:22, 10.10.10.10:6443 och 10.10.10.21:50000 lyckades men ingen data levererades (tunnelyckan accepterar men ingen connector svarar). Twingate rapporterade status online och resursen 'Datacenter datorer' (10.10.10.0/24) med giltig auth, sa det var INTE en auth- eller konfigurationsfraga - anslutaren som serverar resursen var nere.

Konsekvens: ingen serverlogg (dnsmasq/matchbox), ingen talosctl, ingen UniFi-API (nyckeln lases ur 1Password och op-sesningen gick ut). pxe/recover-worker.sh kunde inte koras, eftersom det behover bade UniFi-nyckeln (power-cycle) och API-atkomst.

Sidauppteckning: 1Password CLI-sessionen (op signin) gick ut och gick inte att fornya automatiskt - desktop-appens CLI-integration maste vara aktiv. op item list for host: 10.1.1.1 gav not
