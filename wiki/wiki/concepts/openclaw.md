---
title: OpenClaw
type: concept
summary: OpenClaw är AI-agent-gatewaysen i klustret - en gateway-pod med PVC, OIDC-skydd via Envoy Gateway och MiniMax som modellleverantör.
tags: [openclaw, ai, gateway, oidc, kubernetes]
sources: [infrastructure-flux/components/openclaw/]
created: 2026-11-27
updated: 2026-11-27
related: [concepts/envoy-gateway, concepts/oidc, notes/openclaw-gateway-refuserar-bind-lan-utan-auth]
---

# OpenClaw

## Definition

OpenClaw är en AI-agent-gateway som körs som ett enda Deployment i namespace
`openclaw` i produktionsklustret. Gatewayen exponerar ett multiplexat port
18789 (WebSocket control/RPC + HTTP-API) och använder MiniMax Coding Plan
(`minimax/MiniMax-M3`, `minimax/MiniMax-M2.7`) som modellleverantör.

## Komponentstruktur

```
infrastructure-flux/components/openclaw/
├── kustomization.yaml   # Kustomize-aggregator
├── namespace.yaml       # Namespace: openclaw
├── pvc.yaml             # 10Gi (Longhorn) - config + workspace
├── configmap.yaml       # openclaw.json + AGENTS.md
├── openclaw-secrets.enc.yaml  # API-nycklar (SOPS)
├── oidc-secret.enc.yaml        # OIDC-klient för Envoy SecurityPolicy
├── service.yaml         # ClusterIP:18789
└── deployment.yaml      # Gateway + init-config (seed-if-missing)
```

Publika vägen: `openclaw.simonbrundin.com` → Cloudflare Tunnel → Envoy Gateway
→ HTTPRoute `openclaw-route` → Service `openclaw:18789` → pod.

## Autentisering

Två lager:

1. **Envoy Gateway** (SecurityPolicy `openclaw-oidc`) autentiserar webbläsaren
   mot ZITADEL och skickar vidare cookie-baserad session.
2. **OpenClaw** kör i `gateway.auth.mode: "trusted-proxy"` och kräver att
   identiteten kommer från en betrodd proxy i headern `x-forwarded-user`
   (satt av HTTPRoutens `RequestHeaderModifier`).

Varför inte `mode: "none"` eller `token`: se
[[notes/openclaw-gateway-refuserar-bind-lan-utan-auth]].

## Config på PVC

`openclaw.json` och `workspace/AGENTS.md` seedas från ConfigMapen **bara om de
saknas** - efter första boot äger PVC:n configen. En ConfigMap-ändring
appliceras alltså inte; antingen patchas filen på PVC:n eller configen
reseedas (se komponentens README).

## Kommandon

```bash
kubectl -n openclaw get pod -l app.kubernetes.io/name=openclaw
kubectl -n openclaw logs deploy/openclaw --tail=50
kubectl -n openclaw port-forward svc/openclaw 18789:18789
```

## Kopplingar

- [[concepts/envoy-gateway]]
- [[concepts/oidc]]
- [[notes/openclaw-gateway-refuserar-bind-lan-utan-auth]]
