---
title: Envoy Gateway
type: concept
summary: Envoy Gateway är en implementation av Kubernetes Gateway API som använder Envoy proxy som control plane och data plane.
tags: [envoy, gateway-api, ingress, reverse-proxy, http-routing]
sources: [infrastructure-flux/components/envoy/]
created: 2026-06-02
updated: 2026-10-24
related: [concepts/http-routing-flow, concepts/oidc, concepts/zitadel]
---

# Envoy Gateway

## Definition

Envoy Gateway är en implementation av Kubernetes Gateway API som använder Envoy proxy som control plane och data plane. Det tillhandahåller ingress-kontroll och avancerad HTTP/TCP-routing för Kubernetes-tjänster.

## Nyckelpunkter

### Installation via Flux (OCI Repository)

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: envoy-gateway
  namespace: envoy
spec:
  interval: 5m
  chartRef:
    kind: OCIRepository
    name: envoy-gateway
    namespace: flux-system
  values:
    deployment:
      envoy:
        service:
          type: NodePort
          ports:
            - name: http
              port: 80
              nodePort: 30080
    config:
      envoyGateway:
        extensionApis:
          enableBackend: true
        provider:
          type: Kubernetes
  install:
    timeout: 5m
  upgrade:
    strategy:
      name: RetryOnFailure
      retryInterval: 5m
```

### Resurser som skapas

| Fil | Typ | Beskrivning |
|-----|-----|-------------|
| namespace.yaml | Namespace | envoy namespace |
| ocirepository.yaml | OCIRepository | Envoy Gateway chart |
| release.yaml | HelmRelease | Helm release |
| gatewayclass.yaml | GatewayClass | Klassen för Gateway |
| gateway.yaml | Gateway | Ingress Gateway |
| reference-grant.yaml | ReferenceGrant | Åtkomst till backend-tjänster |

### Gateway API Resources

#### GatewayClass

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: envoy
spec:
  controllerName: gateway.envoy.io/gateway-controller
```

#### Gateway

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: envoy
  namespace: envoy
spec:
  gatewayClassName: envoy
  listeners:
    - name: http
      port: 80
      protocol: HTTP
      allowedRoutes:
        namespaces:
          from: All
    - name: https
      port: 443
      protocol: HTTPS
      allowedRoutes:
        namespaces:
          from: All
```

#### HTTPRoute

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: my-app
  namespace: my-app
spec:
  parentRefs:
    - name: envoy
      namespace: envoy
  hostnames:
    - myapp.simonbrundin.com
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /
      backendRefs:
        - name: my-app-service
          port: 8080
```

### Gateway API vs Ingress

| Feature | Ingress | Gateway API |
|---------|---------|-------------|
| **Traffic management** | Basic routing | L7 routing,负载均衡 |
| **Protocol support** | HTTP/HTTPS | HTTP, HTTPS, TCP, TLS |
| **Filter chains** | Limited | Native support |
| **Weighted routing** | With annotations | Native |
| **Header manipulation** | Limited | Full support |
| **Backend validation** | No | Yes (ReferenceGrant) |

### Flöde: HTTP Request

```
Klient → DNS → External IP (MetalLB) → Gateway (Envoy)
                                            ↓
                                      HTTPRoute match
                                            ↓
                                      Backend Service
                                            ↓
                                      Pod
```

### Konfiguration

**Envoy-extensions aktiverade:**
- `enableBackend: true` - Stöd för att ropa på backend-tjänster

**Provider:** Kubernetes - Envoy konfigurerar sig via K8s resources

### Kommandon för felsökning

```bash
# Lista GatewayClasses
kubectl get gatewayclass

# Lista Gateways
kubectl get gateway -A

# Lista HTTPRoutes
kubectl get httproute -A

# Visa Envoy-config
kubectl get configmap -n envoy envoy-config -o yaml

# Logs från Envoy Gateway
kubectl logs -n envoy -l control-plane=envoy-gateway

# Logs från Envoy proxy
kubectl logs -n envoy -l app=envoy
```

### Vanliga problem

| Problem | Orsak | Lösning |
|---------|-------|---------|
| 404 vid åtkomst | HTTPRoute inte linked till Gateway | Verifiera parentRefs |
| 503 Service Unavailable | Backend-tjänst nere | Kontrollera backend-pod hälsa |
| Certifikatfel | ReferenceGrant saknas | Skapa ReferenceGrant för namnområdet |
| Timeout | Fel port eller target i backendRef | Verifiera service och port |

### Ingress med cert-manager

Envoy Gateway integrerar med cert-manager för automatisk certifikathantering:

1. **HTTP → HTTPS redirect** kan konfigureras
2. **TLS termination** sker i Envoy
3. **Certificate CRD** skapas av cert-manager

## Kopplingar

- [[entities/kubernetes]]
- [[concepts/metallb]]
- [[concepts/cert-manager]]

## Relaterade

- [[concepts/http-routing-flow]]
- [[concepts/oidc]]
- [[concepts/zitadel]]
