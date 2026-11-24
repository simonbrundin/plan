---
title: PXE-status fungerar i produktionsklustret - TFTP svarar inte fran pod-natet
type: note
tags: [pxe, dashboard, produktion, nätverk, session, projekt:tmp]
created: 2026-11-03
updated: 2026-11-03
status: active
related: [entities/pxe-server, notes/pxe-status-pa-dashboard-startsidan-tftp-gar-inte-att-kontrollera-fran-pod-natet]
---

Verifierat mot det riktiga produktionsklustret (Talos, controlplane-1..3 pa 10.10.10.11-13, pod-CIDR 10.244.x) med en testpod:

- Matchbox HTTP (10.10.10.2:8080) nar ALLTID korrekt fran pod-natet: /boot.ipxe, /ipxe?mac=..., /assets/talos-config/nodes/worker-1.yaml och /assets/rock4c-plus/kernel-arm64 svarade 200/OK fran en pod i cluster ns default.
- TFTP (UDP 69) fran samma pod: socketen gick open mot 10.10.10.2:69 men RRQ-svaret kom aldrig tillbaka (0 byte). Detta ar samma beteende som i dev och anledningen till att TFTP-checken klassas som 'unknown' (Okand) och aldrig 'fail'.
- Nodkontrollerna i API:t anvander HEAD mot matchbox. Det fungerar (HTTP 200), sa ingen anpassning behovs.

Viktigt for produktionsdrift: TFTP-kontrollen tar 3 extra sekunder per anrop och kommer alltid visa 'Okand'. Vill man slippa det sattas NUXT_PXE_TFTP_ENABLED=false (stods bade "false" och boolean false).

Bygget verifierat: 'nub run build' går igenom och routen hamnar i .output/server/chunks/routes/api/pxe-status.get.mjs. En byggd server (node .output/server/index.mjs) svarade korrekt pa /api/pxe-status med ok=true och fyra 'ok' plus en 'unknown'.

Kvar att verifiera nar det finns en live-produktionsdeployment: vilket kluster som egentligen ligger bakom dashboard.simonbrundin.com. Vid kontroll var namespaces 'dashboard' och 'envoy' inte deployade i Talos-klustret (bara kube-system fanns), och gatewayn pa 10.10.10.101 slog inte TLS-handshake. Nagot i produktionsinstallationen behover ses over - men naretskillnaden ar densamma i de bada klustren, sa funktionen i sig ar inte problemet.

## Relaterade

- [[notes/pxe-status-pa-dashboard-startsidan-tftp-gar-inte-att-kontrollera-fran-pod-natet]]
- [[entities/pxe-server]]
