---
title: Sammanfattning: Infrastructure Flux
type: summary
summary: `infrastructure-flux/README.md` — Arkitektur och konfigurationsguide.
tags: [flux, gitops, kustomize, kubernetes]
sources: infrastructure-flux/README.md
created: 2026-06-02
updated: 2026-10-24
---

# Sammanfattning: Infrastructure Flux

## Källa

`infrastructure-flux/README.md` — Arkitektur och konfigurationsguide.

## Huvudpunkter

### FluxOperator vs Vanilla Flux

| Aspekt | FluxOperator | Vanilla Flux |
|--------|--------------|--------------|
| Config | `FluxInstance` CRD | `GitRepository` + `Kustomization` |
| Installation | Helm chart | flux bootstrap cli |
| Auto-sync | Ja | Nej |

### Flöde

```
FluxInstance → Skapar intern Kustomization → path: infrastructure-flux
                                               ↓
                                         kustomize build
                                               ↓
                                         components/*/
```

### Lägga till ny component

1. `mkdir infrastructure-flux/components/<namn>/`
2. Skapa: `namespace.yaml`, `source.yaml`, `release.yaml`, `kustomization.yaml`
3. Uppdatera `components/kustomization.yaml`
4. Commit & push

## Relaterade

- [[concepts/flux-operator]]
- [[concepts/gitops]]
