---
title: Cloudflare Tunnel aktiverat - plan.simonbrundin.com publikt nåbar
type: note
tags: [cloudflare, tunnel, ingress, dns, projekt:infrastructure]
---

Läge: publik IPv4 är död (Starlink CGNAT), så Cloudflare Tunnel aktiverades som publik ingress.

Vad som gjordes 2026-11-26 (commit 1060788e):
1. Komponenten cloudflared toggs fram ur kommentar i components/kustomization.yaml.
2. HelmRelease omskriven till lokalt hanterad tunnel: chart cloudflare-tunnel 0.3.2, credentials.json (AccountTag eacb088cc62989f84dcd8788a62d5ed, TunnelID 6a223f33-98f4-4075-8743-edabf3213773) i SOPS-secreten cloudflared-token. Den gamla TUNNEL_TOKEN-secretens nyckelformat passade inte charten (den vill ha credentials.json).
3. Ingress: plan.simonbrundin.com -> http://envoy-envoy-simons-gateway-2efb778c.envoy.svc.cluster.local:80 sa att befintliga HTTPRoutes (/api/v1 -> go-api) fortsatter gälla.
4. DNS: CNAME plan.simonbrundin.com -> 6a223f33-98f4-4075-8743-edabf3213773.cfargotunnel.com (proxied), skapad med cert-manager-tokenen (DNS-skrivrättigheter räcker).

Verifierat: https://plan.simonbrundin.com -> 200 (giltigt CF-cert), /api/v1/* routas till go-api (GIN-logg visar CF-edge-IP 31.209.41.238). Tunneln registrerade 4 QUIC-anslutningar mot arn (Stockholm).

Fallgropar: (1) cert-manager-tokenen har INTE Cloudflare Tunnel:Edit - lista/skapa tunnel via API ger 7003; lokalt hanterad config är därför rätt val. (2) Lokal dnsmasq på arbetsstationen override:ar *.simonbrundin.com -> 10.10.10.101, så testa publikt med dig @1.1.1.1 + curl --resolve. (3) Fler hostnames läggs till genom nya ingress-regler i release.yaml + CNAME.
