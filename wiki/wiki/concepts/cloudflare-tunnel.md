---
title: Cloudflare Tunnel (cloudflared)
type: concept
summary: Cloudflare Tunnel är utkommenterat i komponentlistan och används inte - trafiken går via DuckDNS och port-forwarding i stället.
tags: [cloudflare, tunnel, cfd, dns, ingress, deprecated]
sources: [infrastructure-flux/components/kustomization.yaml, infrastructure-flux/components/cloudflared/]
created: 2026-06-02
updated: 2026-10-24
status: deprecated
---

# Cloudflare Tunnel (cloudflared)

## Status: deprecated

Cloudflare Tunnel är **utkommenterat** i
`infrastructure-flux/components/kustomization.yaml`:

```yaml
  # - cloudflared # Ev lösning ifall jag inte får till en fast publik ipadress
```

Komponenten finns kvar i repot som ett alternativ, men den deployas
inte. Det betyder att trafiken till klustret går via **DuckDNS +
port-forwarding i hemroutern**, inte via en tunnel - se
[[concepts/http-routing-flow]].

Tunnelen aktiveras om det blir aktuellt med en fast publik IP-adress.

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
