---
title: MetalLB
type: concept
summary: MetalLB är en load-balancer-implementation för Kubernetes som fungerar på layer 2 (L2) i nätverksstacken.
tags: [metallb, load-balancer, networking, l2, ip-address-pool]
sources: [infrastructure-flux/components/metallb/]
created: 2026-06-02
updated: 2026-10-24
---

# MetalLB

## Definition

MetalLB är en load-balancer-implementation för Kubernetes som fungerar på layer 2 (L2) i nätverksstacken. Den tilldelar externa IP-adresser till tjänster av typen `LoadBalancer` när ingen extern load-balancer finns (t.ex. i bare-metal miljöer).

## Nyckelpunkter

### Installation via Flux

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: metallb
  namespace: metallb-system
spec:
  chart:
    spec:
      chart: metallb
      version: "0.14.5"
      sourceRef:
        kind: HelmRepository
        name: metallb
  install:
    timeout: 15m
  upgrade:
    timeout: 15m
```

### IP-adresspool

Poolen ligger i ett eget komponent-register,
`infrastructure-flux/components/metallb-address-resources/`, så att
den kan appliceras oberoende av HelmRelease:

```yaml
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: default-pool
  namespace: metallb-system
spec:
  addresses:
    - 10.10.10.101-10.10.10.250
```

### L2 Advertisement

```yaml
apiVersion: metallb.io/v1beta1
kind: L2Advertisement
metadata:
  name: default
  namespace: metallb-system
spec:
  ipAddressPools:
    - default-pool
```

### Användning

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  type: LoadBalancer
  ports:
    - port: 80
      targetPort: 8080
  selector:
    app: my-app
```

När tjänsten skapas tilldelar MetalLB en IP från `default-pool` och announcerar den via ARP/NDP.

### Flöde: LoadBalancer-tjänst

```
1. Service skapas med type: LoadBalancer
       ↓
2. Kubernetes API begär extern IP från MetalLB
       ↓
3. MetalLB allokerar IP från address pool (10.10.10.101-250)
       ↓
4. MetalLB announcer IP:t via L2 (ARP/NDP)
       ↓
5. Klienter kan nu nå tjänsten på den externa IP:n
```

### L2 Mode (ARP/NDP)

**Fördelar:**
- Enkel konfiguration
- Fungerar i de flesta nätverk
- Ingen ändring av nätverksutrustning krävs

**Nackdelar:**
- Ingen lastbalansering mellan noder (en nod tar emot all trafik)
- Skalbarhetsbegränsningar vid många samtidiga tjänster

### Kommandon för felsökning

```bash
# Visa alla address pools
kubectl get ipaddresspool -A

# Visa L2 advertisements
kubectl get l2advertisement -A

# Visa metallb speaker pods
kubectl get pods -n metallb-system -l app=metallb,component=speaker

# Kontrollera om IP är allokerad
kubectl get svc -A | grep LoadBalancer

# Logs från controller
kubectl logs -n metallb-system -l app=metallb,component=controller
```

### Vanliga problem

| Problem | Orsak | Lösning |
|---------|-------|---------|
| `<pending>` på LoadBalancer | MetalLB inte installerat eller IP-pool tom | Verifiera MetalLB och IP pool |
| IP utanför pool | Mis-match mellan pool och advertisement | Koppla pool till L2Advertisement |
| Dubbla ARP-responser | Flera noder announcer samma IP | Kontrollera L2Advertisement config |
| IP fungerar inte | Brandvägg blockerar | Öppna för ARP/NDP på switchar |

## Kopplingar

- [[entities/kubernetes]]
- [[concepts/envoy-gateway]]
