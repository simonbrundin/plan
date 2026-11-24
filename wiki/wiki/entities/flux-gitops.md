---
title: Flux GitOps
type: entity
summary: GitOps-operatorn som synkar klustrets tillstånd från git; här styrd av FluxOperator.
tags: [flux, gitops, infrastructure, kubernetes]
sources: [infrastructure-flux/README.md, wiki/concepts/FluxOperator.md, wiki/concepts/GitOps.md]
created: 2026-06-02
updated: 2026-10-24
status: active
related: [entities/kubernetes, summaries/flux-components]
---

## Sammanfattning

Flux är en GitOps-operator för Kubernetes som automatiskt syncar klustrets state till git. I detta repo används **FluxOperator** (`fluxcd.controlplane.io`) som hanterar all konfiguration via `FluxInstance` CRD:er.

Flux är en GitOps-operator för Kubernetes som automatiskt synkar konfiguration från git till klustret. Ändringar görs i git och Flux applicerar dem automatiskt — ingen manuell `kubectl apply`.

## Detaljer

### FluxOperator

FluxOperator läser `FluxInstance` och skapar resurser automatiskt:

```yaml
spec:
  sync:
    kind: GitRepository
    url: "https://github.com/simonbrundin/infrastructure"
    ref: "refs/heads/main"
    path: "infrastructure-flux"
```

### Flöde

```
FluxInstance → Skapar intern GitRepository → Skapar Kustomization
                                              ↓
                                        kustomize build
                                              ↓
                                        components/*/
                                                ↓
                                        cert-manager/, longhorn-system/, etc.
```

### Flux-komponenter i klustret

| Komponent | Ansvar |
|-----------|--------|
| source-controller | Läser git repositories, Helm repos |
| kustomize-controller | Applicerar kustomize builds |
| helm-controller | Hanterar HelmReleases |
| notification-controller | Hanterar alerts |

### SOPS-decryption

Flux kustomize-controller decryptar krypterade secrets via:

```yaml
spec:
  decryption:
    provider: sops
    secretRef:
      name: sops-age
```

## Katalogstruktur

```
infrastructure-flux/
├── kustomization.yaml              # Root
├── components/
│   ├── kustomization.yaml          # Aggregator
│   ├── cert-manager/               # Exempel component
│   ├── longhorn-system/           # Storage
│   └── secrets/                    # SOPS-krypterade secrets
```

## Kommandon för felsökning

```bash
# Lista alla Flux-resurser
flux get all

# Visa GitRepository status
flux get sources git

# Visa Kustomization status
flux get kustomizations

# Visa HelmReleases
flux get hr
```

## Relaterade

- [[concepts/flux-operator]]
- [[concepts/gitops]]
- [[concepts/sops]]
- [[entities/kubernetes]]
- [[entities/talos-linux]]
- [[summaries/flux-components]]

## Arkitektur
### Katalogstruktur
```
infrastructure/infrastructure-flux/
├── kustomization.yaml              # Root Kustomize config
├── components/                      # Återanvändbara components
│   ├── cert-manager/
│   ├── cloudnative-pg/
│   ├── flux/
│   ├── kube-prometheus-stack/
│   ├── longhorn-system/
│   ├── metallb/
│   └── secrets/                     # SOPS-krypterade hemligheter
└── clusters/production/
```
### Component-mönster
Varje component följer samma struktur:
- `namespace.yaml` — Namespace resource
- `source.yaml` — HelmRepository (definierar var helm-charten finns)
- `release.yaml` — HelmRelease (installerar charten)
- `kustomization.yaml` — Aggregator för componenten
### Installerade components
| Component | Beskrivning |
|-----------|-------------|
| cert-manager | TLS-certifikat med Let's Encrypt |
| cloudnative-pg | PostgreSQL operator |
| flux | FluxOperator & konfiguration |
| kube-prometheus-stack | Monitoring & alerting |
| longhorn-system | Block storage |
| metallb | Lastbalansering |
| secrets | SOPS-hemligheter |
| zitadel | Identity provider (OIDC) |
| auth | OIDC-baserade ClusterRoles |
| kubelet-csr-approver | Auto-godkännande av CSR:er |
| envoy | Gateway API-implementation |
| kyverno | Policy-engine |
| dot-ai | AI-gateway |
| agentgateway | AI-agent-gateway |
| openclaw | Agentdeployment |
| actions-runner-controller | Self-hosted GitHub-runners |
## Kommandon
```bash
flux get all   # Visa status för alla Flux-resurser
```
