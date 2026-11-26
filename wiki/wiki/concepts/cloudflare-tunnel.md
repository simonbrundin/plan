---
title: Cloudflare Tunnel (cloudflared)
type: concept
summary: Cloudflare Tunnel är aktiverat (2026-11-26) och serverar plan.simonbrundin.com publikt - den enda fungerande publika ingressen sedan Starlink-CGNAT.
tags: [cloudflare, tunnel, cfd, dns, ingress]
sources: [infrastructure-flux/components/kustomization.yaml, infrastructure-flux/components/cloudflared/]
created: 2026-06-02
updated: 2026-11-26
status: active
---

# Cloudflare Tunnel (cloudflared)

## Status: aktiverad 2026-11-26

Cloudflare Tunnel är **aktiverat** i
`infrastructure-flux/components/kustomization.yaml` och är den enda
fungerande publika ingressen sedan Starlink-bytet (CGNAT, ingen publik
IPv4 - se [[notes/ipv6-ingress-till-envoy-designlage-2026-11-15]]).

- Lokalt hanterad tunnel (credentials.json i SOPS-secreten
  `cloudflared-token`), tunnel-ID `6a223f33-98f4-4075-8743-edabf3213773`,
  account `eacb088cc62989f84dcd8788a62d5ed`.
- Ingress i `components/cloudflared/release.yaml`:
  `plan.simonbrundin.com` → `http://envoy-envoy-simons-gateway-2efb778c.envoy.svc.cluster.local:80`
  (Envoy routar vidare via befintliga HTTPRoutes, t.ex. `/api/v1` → go-api).
- DNS: CNAME `plan.simonbrundin.com` → `<tunnel-id>.cfargotunnel.com`
  (proxied). Wildcard-CNAME → krokviken.duckdns.org är kvar (DNS only)
  men är död för inkommande trafik.
- Flöde: användare → Cloudflare edge (TLS) → QUIC-tunnel → cloudflared-pod
  → Envoy → service. Verifierat 2026-11-26: `https://plan.simonbrundin.com` → 200.
- Cert-manager-API-tokenen har bara DNS-rättigheter - tunnelkonfiguration
  via API kräver ett token med Cloudflare Tunnel:Edit.

## Definition

Cloudflare Tunnel (tidigare Argo Tunnel) skapar en säker tunnel mellan
interna tjänster och Cloudflares nätverk. Det möjliggör åtkomst till
Kubernetes-tjänster utan att öppna externa portar i brandväggen.

## Nyckelpunkter

### Installation via Flux

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: cloudflared
  namespace: cloudflared
spec:
  chart:
    spec:
      chart: cloudflare-tunnel
      version: ">=0.3.0"
      sourceRef:
        kind: HelmRepository
        name: cloudflared
```

### Resurser

| Fil | Typ | Beskrivning |
|-----|-----|-------------|
| namespace.yaml | Namespace | cloudflared namespace |
| source.yaml | HelmRepository | Cloudflare chart repo |
| release.yaml | HelmRelease | Helm release |
| credentials.enc.yaml | Secret (SOPS) | Tunnel-credentials |
| config.yaml | ConfigMap | Tunnel-konfiguration |

Tunneln skulle proxa trafik till:

| Tjänst | Internt target |
|--------|----------------|
| Grafana | kube-prometheus-stack-grafana:443 |
| Kubernetes API | kubernetes.default.svc:443 |

## Felsökning

| Problem | Orsak | Lösning |
|---------|------|---------|
| `Tunnel connection failed` | Fel tunnel-ID eller credentials | Verifiera credentials.json |
| `Cannot connect to origin` | Backend-tjänst nere | Kontrollera tjänstens hälsa |
| DNS visas inte | CNAME inte skapad | Skapa CNAME i Cloudflare DNS |
| `Invalid token` | Token är expired | Generera ny token |

## Kopplingar

- [[concepts/http-routing-flow]]
- [[entities/kubernetes]]
- [[concepts/sops]]
