---
title: FluxOperator
type: concept
summary: Operator som installerar och styr Flux via FluxInstance-CRD istället för flux bootstrap.
tags: [flux, flux-operator, gitops, installation, kubernetes, operator]
sources: [infrastructure-flux/components/flux/flux-instance.yaml, infrastructure-flux/components/flux/security-policy.yaml]
created: 2026-06-02
updated: 2026-10-24
status: active
related: [entities/kubernetes, summaries/flux-components, summaries/infrastructure-flux]
---

# FluxOperator

## Definition

FluxOperator (`fluxcd.controlplane.io`) är en Kubernetes-operator som
installerar och hanterar Flux utan traditionell CLI-bootstrap.
Istället läser den en `FluxInstance` CRD och skapar resurserna själv.

FluxOperator är den enda synkvägen i det här klustret - både
komponenter och applikationer hanteras av den.

## Skillnad mot vanilla Flux

| Aspekt | FluxOperator | Vanilla Flux |
|--------|--------------|--------------|
| Config | `FluxInstance` CRD | `GitRepository` + `Kustomization` |
| Installation | Helm chart + CRD | `flux bootstrap` |
| GitRepository | Skapas automatiskt | Skapas manuellt i git |
| Auto-sync | Ja | Kräver manuell konfiguration |

## FluxInstance

`infrastructure-flux/components/flux/flux-instance.yaml`:

```yaml
apiVersion: fluxcd.controlplane.io/v1
kind: FluxInstance
metadata:
  name: flux
  namespace: flux-system
spec:
  distribution:
    version: "2.x"
    registry: "ghcr.io/fluxcd"
    artifact: "oci://ghcr.io/controlplaneio-fluxcd/flux-operator-manifests"
  components:
    - source-controller
    - source-watcher
    - kustomize-controller
    - helm-controller
    - notification-controller
  cluster:
    type: kubernetes
    size: medium
    multitenant: false
    networkPolicy: true
    domain: "cluster.local"
  sync:
    kind: GitRepository
    url: "ssh://git@github.com/simonbrundin/infrastructure"
    ref: "refs/heads/main"
    path: "infrastructure-flux"
    pullSecret: "flux-git-ssh-key"
```

### SOPS-dekryptering

`spec.kustomize.patches` lägger till SOPS-provider på den genererade
Kustomizationen:

```yaml
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

Utan den patchen bryts varje krypterat secret i klustret - se
[[concepts/sops]].

## Internt flöde

```
FluxInstance → Skapar intern GitRepository → Skapar Kustomization
                                                ↓
                                          path: infrastructure-flux
                                                ↓
                                          kustomize build
                                                ↓
                                          components/*/ och apps/*/
```

Operatorn läser **inte** `gitrepository.yaml` från git - den skapar sin
egen internt.

### Kustomization-hierarki

```
FluxInstance (FluxOperator)
  └── Skapar Kustomization
        └── path: infrastructure-flux
              └── kustomize build
                    ├── components/   # Operatörer och plattformskomponenter
                    └── apps/        # Pekare till externa repos
```

## Katalogstruktur

```
infrastructure-flux/
├── kustomization.yaml              # Root: components + apps
├── components/
│   ├── kustomization.yaml          # Aggregator - varje komponent registreras här
│   ├── flux/                       # FluxOperator, OIDC, security policy
│   ├── secrets/                    # SOPS-nycklar
│   └── <component>/
│       ├── namespace.yaml
│       ├── source.yaml             # HelmRepository/OCIRepository
│       ├── release.yaml            # HelmRelease
│       └── kustomization.yaml
├── apps/
│   └── <app>/                      # GitRepository + Flux Kustomization
└── clusters/production/
    └── flux-system/                # Dokumentation, appliceras inte
```

## Lägga till en ny komponent

1. Skapa mappen `infrastructure-flux/components/<namn>/`
2. Skapa filerna:
   - `namespace.yaml` - Namespace
   - `source.yaml` - HelmRepository/OCIRepository
   - `release.yaml` - HelmRelease
   - `kustomization.yaml` - aggregator
3. **Registrera komponenten i `components/kustomization.yaml`** - utan
   det deployas den aldrig
4. Validera lokalt: `kustomize build infrastructure-flux/components/<namn>`
5. Committa och pusha - FluxOperator synkar automatiskt

## Lokal validering

```bash
# En komponent
kustomize build infrastructure-flux/components/<component> > /dev/null

# Hela trädet
kustomize build infrastructure-flux > /dev/null
```

## Felsökning

```bash
# FluxInstance-status
kubectl -n flux-system get fluxinstance flux

# Synkstatus
flux get all

# Reconcilera manuellt
flux reconcile kustomization <component> --with-source
```

## Kopplingar

- [[concepts/gitops]]
- [[concepts/sops]]
- [[entities/flux-gitops]]
- [[entities/kubernetes]]
- [[summaries/flux-components]]
- [[summaries/infrastructure-flux]]
