---
title: HTTP Routing Flow
type: concept
summary: Hur trafik flödar från internet till tjänster i Kubernetes-klustret via DNS, lastbalansering och gateway-routing.
tags: [networking, routing, dns, cloudflare, duckdns, kubernetes, envoys-gateway]
created: 2026-06-03
updated: 2026-10-24
---

# HTTP Routing Flow

## Definition

Hur trafik flödar från internet till tjänster i Kubernetes-klustret via DNS, lastbalansering och gateway-routing.

## Flödesdiagram

```
Internet
    │
    ▼
Cloudflare DNS
*.simonbrundin.com ──→ krokviken.duckdns.org (CNAME, DNS-only)
    │
    ▼
DuckDNS
krokviken.duckdns.org ──→ 62.108.213.168 (din publika WAN-IP)
    │
    ▼
Hemrouter
Port forwarding: 443 → 10.10.10.101
    │
    ▼
MetalLB LoadBalancer
10.10.10.101 (Envoy Gateway)
    │
    ▼
Envoy Gateway
Läser HTTPRoute → vidarebefordrar till tjänst
    │
    ▼
Kubernetes-tjänst
kube-prometheus-stack-grafana:80
    │
    ▼
Grafana Pod
```

## DNS-kedja

| Företag | Domän | Typ | Proxy |
|----------|-------|-----|-------|
| Cloudflare | `*.simonbrundin.com` | CNAME → krokviken.duckdns.org | **DNS only** |
| DuckDNS | `krokviken.duckdns.org` | A → 62.108.213.168 | - |

### Viktigt: Cloudflare Proxy måste vara AV

Orange moln = Proxy aktiverad (tunnel genom Cloudflare)
Grått moln = DNS only (vanlig DNS-resolvering)

Om Cloudflare proxy är på, kommer trafiken inte nå din server.

## Komponenter

### 1. DNS-upplösning

DuckDNS uppdaterar automatiskt din publika IP var 5:e minut via cron-jobb eller token.

### 2. Port Forwarding (Router)

```
Internet:443 → Router → 10.10.10.101:443
```

### 3. MetalLB

Tilldelar `10.10.10.101` till Envoy Gateway som en LoadBalancer IP.

### 4. Envoy Gateway

Tar emot HTTPS-trafik på port 443, terminerar TLS med Let's Encrypt-certifikat.

### 5. HTTPRoute (Kubernetes)

Definierar routing-regler:

```yaml
kind: HTTPRoute
spec:
  hostnames:
    - "grafana.simonbrundin.com"
  rules:
    - backendRefs:
        - kind: Service
          name: kube-prometheus-stack-grafana
          port: 80
```

## TLS-certifikat

Certifikatet hanteras av [[concepts/cert-manager]]:

- **Issuer**: `cloudflare-clusterissuer` (ClusterIssuer)
- **Secret**: `simonbrundin` i `cert-manager` namespace
- **DNS**: `*.simonbrundin.com`, `simonbrundin.com`
- **Giltigt**: Let's Encrypt

## Felsökning

### Steg-för-steg DNS- och routing-tester

Dessa kommandon verifierar varje steg i kedjan:

```bash
# Steg 1: Intern MetalLB IP (direkt på servern/nätverket)
# Adress: 10.10.10.101 = MetalLB VIP tilldelad Envoy Gateway av MetalLB
curl -sI --connect-timeout 5 -H "Host: grafana.simonbrundin.com" http://10.10.10.101
# Förväntat: HTTP/1.1 302 Found → /login
# Denna test verifierar: MetalLB → Envoy → HTTPRoute → Grafana

# Steg 2: Publika WAN-IP (efter port forwarding)
# Adress: 62.108.213.168 = Din publika IP som DuckDNS pekar mot
curl -sI --connect-timeout 5 -H "Host: grafana.simonbrundin.com" http://62.108.213.168
# Förväntat: HTTP/1.1 302 Found → /login
# Denna test verifierar: Router port forwarding → 10.10.10.101

# Steg 3: DuckDNS-domän (DNS-upplösning)
# Adress: krokviken.duckdns.org = DuckDNS-pekar mot 62.108.213.168
curl -sI --connect-timeout 5 -H "Host: grafana.simonbrundin.com" http://krokviken.duckdns.org
# Förväntat: HTTP/1.1 302 Found → /login
# Denna test verifierar: DuckDNS DNS-upplösning → publika IP

# Steg 4: Cloudflare (fullständig kedja med HTTPS)
# Adress: grafana.simonbrundin.com = CNAME → krokviken.duckdns.org
curl -sI --connect-timeout 5 -H "Host: grafana.simonbrundin.com" https://grafana.simonbrundin.com
# Förväntat: HTTP/2 302 Found → /login
# Denna test verifierar: Cloudflare → DuckDNS → Publika IP → Router → MetalLB → Envoy → Grafana
```

### Adresser förklarade

| Adress | Typ | Vad det är |
|--------|-----|------------|
| `10.10.10.101` | Intern VIP | MetalLB LoadBalancer IP, tilldelad Envoy Gateway |
| `62.108.213.168` | Publik IP | Din WAN-IP, uppdaterad av DuckDNS var 5:e minut |
| `krokviken.duckdns.org` | DuckDNS | pekar mot publika IP:n |
| `grafana.simonbrundin.com` | Cloudflare | CNAME → krokviken.duckdns.org (DNS only) |

### Host-header

`-H "Host: grafana.simonbrundin.com"` behövs eftersom Envoy Gateway använder HTTP Routing baserat på `Host`-headern. Utan denna header kan Envoy inte matcha rätt HTTPRoute.

### Kontrollera Gateway

```bash
kubectl get gateway -n envoy
kubectl get httproute -A
kubectl describe httproute grafana-route -n kube-prometheus-stack
```

### Om något steg misslyckas

| Steg som misslyckas | Möjlig orsak |
|---------------------|--------------|
| Steg 1 (10.10.10.101) | MetalLB, Envoy Gateway, eller HTTPRoute felkonfigurerad |
| Steg 2 (62.108.x.x) | Port forwarding felaktig eller inaktiverad |
| Steg 3 (krokviken.duckdns.org) | DuckDNS uppdaterar inte publika IP:n korrekt |
| Steg 4 (grafana.simonbrundin.com) | Cloudflare proxy aktiverad (ska vara DNS only) |

## Relaterade

- [[concepts/envoy-gateway]] — Gateway API-implementation
- [[concepts/cert-manager]] — TLS-certifikathantering
- [[concepts/metallb]] — LoadBalancer IP-allokering
- [[concepts/cloudflare-tunnel]] — Alternativ tunnel-lösning
