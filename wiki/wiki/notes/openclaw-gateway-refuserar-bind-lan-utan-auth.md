---
title: OpenClaw vägrar binda till LAN utan auth - använd trusted-proxy via Envoy
type: note
summary: OpenClaw crashloopar med "Refusing to bind gateway to lan without auth" eftersom auth.mode=none inte räcker. Lösningen är auth.mode=trusted-proxy plus att Envoy sätter x-forwarded-user.
tags: [openclaw, envoy-gateway, oidc, bug, projekt:simon]
created: 2026-11-27
updated: 2026-11-27
status: active
related: [concepts/openclaw, concepts/envoy-gateway]
---

# OpenClaw vägrar binda till LAN utan auth - använd trusted-proxy via Envoy

OpenClaw-podden i namespace `openclaw` crashloopade (153 restarts) med:

```
[gateway] auth mode=none explicitly configured; all gateway connections are unauthenticated.
Refusing to bind gateway to lan without auth.
Set gateway.auth.token/password (or OPENCLAW_GATEWAY_TOKEN/OPENCLAW_GATEWAY_PASSWORD) or pass --token/--password.
```

Uppströms gav `upstream connect error ... Connection refused` via Envoy.

## Varför

OpenClaw har en fail-closed-skydd: gatewayen får inte lyssna på LAN
(`bind: "lan"` → `0.0.0.0`) utan en auth-väg. `gateway.auth.mode: "none"` är
**inte** en auth-väg - det betyder bara "ingen autentisering". Kombinationen
`bind: lan` + `mode: none` startar därför aldrig i nyare versioner.

## Lösning: trusted-proxy

Eftersom Envoy Gateway redan autentiserar användaren med OIDC ska gatewayn
lita på identitetsheadern från proxyn:

```json
"gateway": {
  "bind": "lan",
  "auth": {
    "mode": "trusted-proxy",
    "trustedProxy": { "userHeader": "x-forwarded-user" }
  },
  "trustedProxies": ["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"]
}
```

Envoy sätter headern i HTTPRouten (Gateway API `RequestHeaderModifier`):

```yaml
filters:
  - type: RequestHeaderModifier
    requestHeaderModifier:
      set:
        - name: x-forwarded-user
          value: simonbrundin@gmail.com
```

## Fällan: token och trusted-proxy är ömsesidigt uteslutande

Med `OPENCLAW_GATEWAY_TOKEN` kvar i deploymenten startar gatewayn inte:

```
Gateway failed to start: gateway auth mode is trusted-proxy, but a shared token is also
configured; remove gateway.auth.token / OPENCLAW_GATEWAY_TOKEN because trusted-proxy and
token auth are mutually exclusive.
```

Env-variabeln måste alltså tas bort ur deploymenten (och port-forward-flödet
med gateway-token dokumenteras om i komponentens README).

## Operativt

`openclaw.json` på PVC:n äger configen efter första boot - ConfigMap-ändringar
 appliceras inte. Patcha filen på PVC:n (t.ex. via en hjälp-pod som mountar
PVC:n) eller använd det dokumenterade reseed-flödet:

```bash
kubectl exec -n openclaw deploy/openclaw -- rm /home/node/.openclaw/openclaw.json
kubectl rollout restart -n openclaw deploy/openclaw
```

## Kopplingar

- [[concepts/openclaw]]
- [[concepts/envoy-gateway]]
