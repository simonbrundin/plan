---
title: Kubernetes
type: entity
summary: Kubernetes (K8s) är containerorkestreringsplattformen i detta kluster.
tags: [kubernetes, k8s, container-orchestration]
sources: [infrastructure-flux/components/kustomization.yaml, infrastructure-flux/apps/kustomization.yaml]
created: 2026-06-02
updated: 2026-11-23
related: [concepts/kubelet-csr-approver, concepts/oidc, concepts/sops, concepts/zitadel, entities/pxe-server, summaries/flux-components, notes/klustret-har-byggts-om-bara-kube-system-flux-och-appar-saknas-2026-11-23]
---

# Kubernetes

## Sammanfattning

Kubernetes (K8s) är containerorkestreringsplattformen i detta kluster. Klustret körs på Talos Linux, och **allt** hanteras av Flux via FluxOperator - komponenter i `components/`, användarapplikationer i `apps/`. ArgoCD används inte längre.

## Aktuellt tillstånd (2026-11-23)

Det här avsnittet beskriver **arkitekturen i `infrastructure-flux/`** (och
[[summaries/flux-components]]). Klustrets *live*-tillstånd har under 2026-11
byggts om och hade vid senaste kontrollen endast `kube-system` - ingen
FluxOperator, inga CRDs, inga appar. Se
[[notes/klustret-har-byggts-om-bara-kube-system-flux-och-appar-saknas-2026-11-23]].

## Arkitektur

### Klusterstruktur

```
┌─────────────────────────────────────────────────────────────┐
│                     Kubernetes Kluster                       │
├─────────────────────────────────────────────────────────────┤
│  Control Plane Nodes          Worker Nodes                  │
│  ┌─────────────────┐        ┌─────────────────┐           │
│  │ kube-apiserver  │        │ kubelet         │           │
│  │ etcd            │        │ kube-proxy      │           │
│  │ kube-scheduler  │        │ flannel/nodelocaldns │      │
│  │ kube-controller │        │ Workload Pods   │           │
│  └─────────────────┘        └─────────────────┘           │
├─────────────────────────────────────────────────────────────┤
│  System Komponenter (FluxOperator, components/)            │
│  cert-manager │ longhorn │ metallb │ prometheus │ etc.    │
├─────────────────────────────────────────────────────────────┤
│  Applikationer (Flux Kustomization, apps/)                 │
│  plan │ dashboard │ foodflow │ homelab │ travelagent      │
└─────────────────────────────────────────────────────────────┘
```

### Nodtyper

| Nodtyp | Antal | Ansvar |
|--------|-------|--------|
| **controlplane** | 3 | Kubernetes API, etcd, scheduler, controller-manager |
| **worker** | 5+ | Applikationsworkloads, storage (Longhorn) |

## Flux-hanterade Komponenter

### Infrastructure-komponenter (infrastructure-flux/components/)

| Komponent | Namespace | Ansvar |
|-----------|-----------|--------|
| [[concepts/flux-operator]] | flux-system | GitOps-operatorn som synkar allt |
| [[concepts/cert-manager]] | cert-manager | TLS-certifikat (Let's Encrypt) |
| [[concepts/longhorn]] | longhorn-system | Persistent storage |
| [[concepts/metallb]] | metallb-system | L2 lastbalansering |
| [[concepts/kube-prometheus]] | kube-prometheus-stack | Monitoring & alerting |
| [[concepts/envoy-gateway]] | envoy | Ingress/Gateway API |
| [[concepts/cloudnative-pg]] | cloudnative-pg | PostgreSQL-operator |
| [[concepts/kyverno-policies]] | kyverno | Policy-engine |
| Kubelet CSR Approver | kubelet-csr-approver | CSR-godkännande |
| Zitadel | zitadel | Identity provider (OIDC) |
| auth | auth | OIDC-baserade ClusterRoles |
| Atlas Operator | atlas-operator | Schemamigreringar |
| dot-ai | dot-ai | AI-gateway |
| Agentgateway | agentgateway | AI-agent-gateway + Postgres |
| Actions Runner Controller | actions-runner-controller | Self-hosted GitHub-runners |
| FreeLLMAPI | freellmapi | LLM-API-tjänst |
| OpenClaw | openclaw | Agentdeployment |
| Secrets | - | SOPS-nycklar för Flux |

### Applikationer (infrastructure-flux/apps/)

Varje app-katalog innehåller en `GitRepository` + `Flux Kustomization` som
drar arbetslasten från ett annat repo.

| App | Beskrivning |
|-----|-------------|
| plan | Plan-appen |
| dashboard | Dashboard |
| foodflow | Foodflow |
| homelab | Hemlabb-applikationer |
| hitster | Hitster |
| spc-displayscreen | Display för serier |
| travelagent | Travelagent |

### CI/CD & Workflow

| Komponent | Beskrivning |
|-----------|-------------|
| actions-runner-controller | Self-hosted GitHub-runners i klustret |
| Flux source-controller | Läser git- och Helm-repositories |
| Flux helm-controller | Hanterar HelmReleases |

## Networking

### Gateway API

Använder Envoy Gateway som implementation av Kubernetes Gateway API:

```yaml
kind: GatewayClass
spec:
  controllerName: gateway.envoy.io/gateway-controller

kind: Gateway
spec:
  gatewayClassName: envoy
  listeners:
    - name: http
      port: 80
      protocol: HTTP
    - name: https
      port: 443
      protocol: HTTPS
```

### Load Balancing

**MetalLB** för L2-balansering:
- IP Pool: `10.10.10.101-10.10.10.250`
- L2 Advertisement för ARP-baserad lastbalansering

### Ingress-rutter

| Applikation | Hostname |
|-------------|----------|
| Grafana | grafana.simonbrundin.com |
| Zitadel | auth.simonbrundin.com |

Gatewayen `simons-gateway` (namespace `envoy`) lyssnar på 80/443 och
hänvisar till Gateway-API-`HTTPRoute` per app. Adressen är
`10.10.10.101` - alltså MetalLB-pools första IP.

## Storage

### Longhorn

Standard storageclass för persistent storage:

```yaml
storageClassName: longhorn
# Egenskaper:
# - Replica: 3
# - Access Mode: RWO (ReadWriteOnce)
# - Volume Binding: Immediate
```

### CloudNativePG

Operator för PostgreSQL-databaser med:
- Automatisk backup
- High availability (Patroni-baserat)
- Connection pooling

## Monitoring & Observability

### Kube Prometheus Stack

Inkluderar:
- Prometheus server
- Alertmanager
- Grafana dashboards
- node-exporter
- kube-state-metrics

### Dashboards

| Dashboard | URL |
|-----------|-----|
| Grafana | grafana.simonbrundin.com |
| Alertmanager | Via Grafana |

## Säkerhet

### Certifikat

**cert-manager** hanterar:
- Let's Encrypt-certifikat
- ClusterIssuer för klusterövergripande cert
- Certificate CRDs för specifika cert

### Secrets

| Metod | Användning |
|-------|------------|
| [[concepts/sops]] | Git-krypterade secrets (Flux-hanterade) - **enda** vägen |

Det finns **ingen Vault och ingen External Secrets Operator** i klustret
(sedan 2026-11-23 verifierat: inga ESO-CRDs, inget vault-namespace).
ESO/Vault-materialet i `infrastructure/old-files/` är ArgoCD-era och
arkiverat. Se [[concepts/sops]] och
[[notes/klustret-anvander-sops-via-flux-inte-vault-eso]].

### Access Control

| Komponent | Beskrivning |
|-----------|-------------|
| Zitadel | OIDC identity provider (self-hosted) |
| auth | ClusterRoles `oidc-viewer`, `oidc-developer`, `oidc-cluster-admin` |
| Kubelet CSR Approver | Automatisk CSR-godkännande |

OIDC-konfigurationen för API-servern ligger i
`talos/patches/controlplane.yaml` (`KubeAuthenticationConfig`) och pekar
på Zitadel-issuern. Rollmappningen sker via Zitadel-roller och
`auth.zitadel.org/role`-annotationen på ClusterRole:arna.

## Relaterade

- [[concepts/kubelet-csr-approver]]
- [[concepts/flux-operator]]
- [[concepts/gitops]]
- [[concepts/oidc]]
- [[concepts/zitadel]]
- [[entities/flux-gitops]]
- [[entities/pxe-server]]
- [[entities/talos-linux]]
- [[summaries/flux-components]]
