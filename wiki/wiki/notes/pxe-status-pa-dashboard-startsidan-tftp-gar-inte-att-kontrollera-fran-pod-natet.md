---
title: PXE-status på dashboard-startsidan: TFTP går inte att kontrollera från pod-nätet
type: note
tags: [pxe, dashboard, feature, nätverk, projekt:tmp]
created: 2026-11-03
updated: 2026-11-17
status: active
related: [entities/pxe-server, notes/pxe-status-fungerar-i-produktionsklustret-tftp-svarar-inte-fran-pod-natet]
---

Dashboardens startsida har nu ett kort 'PXE-server' (HomePxeStatus.vue + /api/pxe-status) som visar om PXE-servern fungerar.

Viktig lärdom om nätverket: PXE-servern (10.10.10.2, matchbox HTTP 8080 + dnsmasq TFTP 69) går att nå via HTTP från dashboardens pod-nät (10.244.x, flannel) - verifierat i både Talos-klustret (10.10.10.11) och kind-dev. Däremot svarar INTE TFTP (UDP 69) från pod-nätet: svarspaketet routas inte tillbaka. Från nodens host-nätverk (10.10.10.11/24, samma L2-segment som PXE-servern) fungerar TFTP och filen laddas ner korrekt.

Därför är TFTP-checken klassad som 'unknown' (Okänd) i API:t, aldrig 'fail'. Endast HTTP-checkerna (iPXE-startskript, MAC-profiluppslagning, nodkonfiguration, bootkärna) avgör om kortet visar 'Fungerar' eller 'Nere'. En rå UDP-probe i dashboarden hade gett ett permanent falskt larm.

Övriga lärdomar:
- Matchbox svarar 200 på HEAD för /boot.ipxe, /ipxe?mac=<mac>, /assets/talos-config/nodes/<nod>.yaml och /assets/rock4c-plus/kernel-arm64.
- Kontrollerna körs parallellt via Promise.all - sekventiell körning tog 25s när servern var nere, nu 8s.
- runtimeConfig: pxeHost (10.10.10.2), pxeHttpPort (8080), pxeTftpPort (69), pxeProbeMac, pxeProbeNode.

## Relaterade

- [[notes/pxe-status-fungerar-i-produktionsklustret-tftp-svarar-inte-fran-pod-natet]]
- [[entities/pxe-server]]
