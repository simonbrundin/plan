---
title: Flannel VXLAN MAC-kollision blockerar controlplane-2 webhooks
type: note
tags: [incident, flux, flannel, networking, projekt:infrastructure]
---

2026-11-26: controlplane-2 och controlplane-3 annonserar och använder samma Flannel VXLAN-MAC e2:92:15:e2:0e:30. På worker-3 pekar enda FDB-posten för MAC-adressen på 10.10.10.13 trots att grannposter finns för båda podsubnäten 10.244.1.0/24 och 10.244.2.0/24. Från controlplane-2 misslyckades TCP till CNPG, Kyverno, Longhorn och cert-manager via pod-IP och service; från controlplane-3 lyckades samma tester. Direkt CNPG- och Longhorn-podanslutning från controlplane-2 fungerar med källan 10.10.10.12 men inte 10.244.1.0 eller 10.244.1.1. MAC-kollisionen är en stark förklaring till felroutad returtrafik och API-serverns webhook-timeouts, men nätverksreparation har ännu inte provats eller verifierats. Bootstrap-healthchecks applicerades från Git och spärrar nu stegen eftersom bland annat agentgateway har Stalled=True. Rot-Flux blockeras fortfarande av Longhorn-webhook. Inga Flannel- eller nodändringar gjordes.
