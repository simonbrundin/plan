---
title: Envoy-VIP 10.10.10.101 stulen av agentgateway-admin - plan.simonbrundin.com onåbar
type: note
tags: [metallb, envoy, gateway, dns, incident, projekt:infrastructure]
created: 2026-11-26
updated: 2026-11-26
status: active
---

Symtom: plan.simonbrundin.com (och övriga *.simonbrundin.com) nås inte trots att klustret, HTTPRoutes, cert och app-pods är friska.

Orsak: agentgateway-proxy-admin (LoadBalancer utan pinad IP i components/agentgateway/admin-service.yaml) autoallokerade 10.10.10.101 - lägsta lediga IP i MetalLB default-pool (10.10.10.101-10.10.10.250). Samma VIP är den som Gateway spec.addresses, dnsmasq-adressen på arbetsstationen och UniFi-portforwarden pekar på. MetalLB vägrade sedan allokera .101 till envoy-tjänsten ('address also in use'), som istället fick 10.10.10.102. .102 annonserades av worker-6 (där Envoy kör) men .101 annonserades av worker-4 (utan Envoy-endpoint, externalTrafficPolicy: Local) - trafik till .101 släpptes därmed på noden.

Fix (2026-11-26, commit 0d751947): pinna agentgateway-admin till 10.10.10.105, lägg till EnvoyProxy med annotation metallb.io/loadBalancerIPs=10.10.10.101 i gateway/envoy-gateway och peka GatewayClass parametersRef dit. Efter controller-restart allokerar MetalLB .101 till envoy-tjänsten och worker-6 annonserar den. Verifierat: curl --resolve plan.simonbrundin.com:443:10.10.10.101 → 200.

Kvar: publik IPv4-väg är fortfarande död (Starlink CGNAT, DuckDNS → 100.67.39.22) - se notes/ipv6-ingress-till-envoy-designlage-2026-11-15. Lokal åtkomst går via split-horizon-DNS: dnsmasq på arbetsstationen (127.0.0.1:53, address=/.simonbrundin.com/10.10.10.101) men systemd-resolved pekade ~simonbrundin.com mot döda 127.0.0.1:5353 och 10.10.10.101:53 - rätta till DNS=127.0.0.1:53 i /etc/systemd/resolved.conf.d/simonbrundin.conf och forward.conf.
