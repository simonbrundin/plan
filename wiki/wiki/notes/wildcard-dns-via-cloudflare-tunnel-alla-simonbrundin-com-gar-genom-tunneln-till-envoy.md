---
title: Wildcard DNS via Cloudflare Tunnel - alla *.simonbrundin.com går genom tunneln till Envoy
type: note
tags: [cloudflare, tunnel, dns, ingress, envoy, projekt:infrastructure]
created: 2026-11-26
updated: 2026-11-26
status: active
---

Slutgiltigt läge 2026-11-26 (commits 26d040ac..810e22f7):

DNS: *.simonbrundin.com -> CNAME 6a223f33-98f4-4075-8743-edabf3213773.cfargotunnel.com (proxied). plan.simonbrundin.com har en explicit CNAME till samma tunnel. mail.simonbrundin.com fick en explicit DNS-only CNAME till krokviken.duckdns.org (SMTP/MX kan inte proxas).

cloudflared: catch-all-regel -> https://envoy-envoy-simons-gateway-2efb778c.envoy.svc.cluster.local:443 med originRequest.originServerName: simonbrundin.com. Envoy routar vidare per Host-header via befintliga HTTPRoutes. Alla 31 route-hostnames svarar nu via tunnelen (200/302/307/500/503 - 500/503 är appproblem, samma svar via LAN-VIP).

Fyra fallgropar som kostade tid:
1. Helm-charten cloudflare-tunnel lägger på en 404-regel EFTER ingressreglerna -> cloudflared vägrar starta ('Rule #1 is matching the hostname'). Lösning: plain manifests (Deployment + ConfigMap) i stället för charten.
2. En ingress-regel utan hostname matchar allt och avvisas; måste skrivas hostname: '*'.
3. Trafik måste gå till Envoys HTTPS-listener (443). Routes med sectionName: https (zitadel-route, freellmapi) ger 404 på http-listenern.
4. originServerName måste ligga under originRequest: (inte toppnivå) och ange apex-domänen - Envoy-certet är wildcard *.simonbrundin.com + apex. Utan det: x509-fel mot k8s-tjänstenamnet.

Kustomize-fälla: configMapGenerator måste ange namespace, annars hamnar configmapen i default-namespace och kustomize skriver inte om referensen i Deploymenten (tyst - poden startar med gammal config). Hash-suffixet på configmapen ger auto-rollout vid configändring.

Verifiering: dig @1.1.1.1 + curl --resolve mot CF-IP (lokal dnsmasq override:ar *.simonbrundin.com -> 10.10.10.101, så vanlig curl mot domännamnet testar LAN-vägen). Se infrastructure-flux/components/cloudflared/README.md.
