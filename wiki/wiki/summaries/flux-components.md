---
title: Sammanfattning: Flux-komponenter
type: summary
summary: `infrastructure-flux/components/` — Alla HelmReleases och HelmRepositories.
tags: [flux, gitops, components, kubernetes]
sources: infrastructure-flux/components/
created: 2026-06-02
updated: 2026-10-24
---

# Sammanfattning: Flux Components

## Källa

`infrastructure-flux/components/` — Alla HelmReleases och HelmRepositories.

## Översikt: Components via FluxOperator

| Komponent | Namespace | Version | Chart Source |
|-----------|-----------|---------|--------------|
| cert-manager | cert-manager | 1.18.2 | jetstack (charts.jetstack.io) |
| kube-prometheus-stack | kube-prometheus-stack | 61.0.0 | prometheus-community |
| metallb | metallb-system | 0.14.5 | metallb |
| longhorn | longhorn-system | 1.10.1 | longhorn (charts.longhorn.io) |
| envoy | envoy | OCI | ghcr.io/ariga/charts |
| cloudnative-pg | cloudnative-pg | 0.26.0 | cloudnative-pg.io/charts |
| kubelet-csr-approver | kubelet-csr-approver | 1.1.0 | postfinance.github.io |
| cloudflared | cloudflared | >=0.3.0 | cloudflare.github.io/helm-charts |
| atlas-operator | atlas-operator | - | OCI: ghcr.io/ariga/charts |
| zitadel | zitadel | 10.0.2 | charts.zitadel.com |
| flux | flux-system | - | FluxInstance CRD |

**Utkommenterade:** `authentik` - ersatt av `zitadel`.

### HelmRepository-typer

| Typ | Användning |
|-----|------------|
| **Vanlig HTTP** | cert-manager, prometheus, metallb, etc. |
| **OCI** | Envoy Gateway, Atlas Operator |

### Common files per component

Alla components har:
- `namespace.yaml` — Namespace resource
- `source.yaml` — HelmRepository/OCIRepository
- `release.yaml` — HelmRelease med chart och values
- `kustomization.yaml` — Component aggregator

### Flux-instance (root config)

```yaml
# flux/flux-instance.yaml
spec:
  sync:
    kind: GitRepository
    url: "ssh://git@github.com/simonbrundin/infrastructure"
    ref: "refs/heads/main"
    path: "infrastructure-flux"
  kustomize:
    patches:
      - patch: |
          - op: add
            path: /spec/decryption
            value:
              provider: sops
              secretRef:
                name: sops-age
        target:
          kind: Kustomization
```

`spec.kustomize.patches` är det som möjliggör SOPS-dekryptering för
alla genererade Kustomizations. Utan den bryts varje krypterat secret
i klustret.

## Relaterade

- [[concepts/flux-operator]]
- [[entities/flux-gitops]]
- [[entities/kubernetes]]
