---
title: GitOps
type: concept
summary: Git är enda sanningskällan för infrastruktur; Flux/FluxOperator applicerar ändringarna till klustret.
tags: [flux, gitops, deployment, infrastructure-as-code, kubernetes]
sources: [AGENTS.md, infrastructure-flux/README.md]
created: 2026-06-02
updated: 2026-11-27
status: active
related: [entities/kubernetes, notes/flux-ager-coredns-configmap-och-longhorn-node-cr-ownership-reglerna-som-gor-det-saker, summaries/infrastructure-flux]
---

## Definition

GitOps är ett operationsramverk där **git är den enda sanna källan**
för infrastruktur- och applikationskonfiguration. Ändringar görs via
git commits, och en operator (här Flux, styrd av FluxOperator) synkar
automatiskt klustret till git-statet.

## Nyckelpunkter

### Princip

1. **Declarative** — All konfiguration är deklarativ (YAML)
2. **Versioned** — Git commits ger historik och rollback
3. **Pull-based** — Operatorn pullar ändringar, ingen push från CI
4. **Automated** — Klustret syncas automatiskt vid git-ändringar

### Verktyg i detta repo

| Verktyg | Typ | Användning |
|---------|-----|------------|
| **FluxOperator** | GitOps operator | Synkar allt i `infrastructure-flux/` |
| **Flux** | GitOps-synkning | source-controller, kustomize-controller, helm-controller |
| **Kustomize** | Config management | Bygger manifests ur `components/` och `apps/` |

**ArgoCD används inte längre.** Allt - både komponenter och
applikationer - hanteras av Flux.

### Kustomize-mönster

```
infrastructure-flux/
├── kustomization.yaml     # rot: components + apps
├── components/            # operatörer och plattformskomponenter
│   ├── kustomization.yaml # lista över varje komponent
│   └── <component>/
│       ├── kustomization.yaml
│       ├── namespace.yaml
│       ├── source.yaml     # HelmRepository/OCIRepository
│       └── release.yaml    # HelmRelease
└── apps/                  # pekare till externa repos
    └── <app>/
        ├── gitrepository.yaml
        └── flux-kustomization.yaml
```

En ny komponent eller app deployas **bara** om den registreras i
aggregatorn (`components/kustomization.yaml` respektive
`apps/kustomization.yaml`) - och det kräver uttryckligt tillstånd.

## Princip

1. **Declare** — Deklarera önskat tillstånd i git
2. **Push** — Committa och pusha ändringar
3. **Sync** — Flux applicerar ändringarna

## Fördelar

| Fördel | Beskrivning |
|--------|-------------|
| Audit trail | Alla ändringar är i git-historik |
| Rollback | Enkel återställning med `git revert` |
| Driftkontroll | Ändringar kräver peer review |
| Reproducerbarhet | Samma config överallt |

## Flux i klustret

Flux är GitOps-operatorn som kör i klustret och bevakar git för
ändringar. Se [[entities/flux-gitops]] och [[concepts/flux-operator]] för
detaljer.

## Kopplingar

- [[concepts/flux-operator]]
- [[entities/flux-gitops]]
- [[entities/kubernetes]]
- [[concepts/sops]]
- [[summaries/infrastructure-flux]]

## Relaterade

- [[notes/flux-ager-coredns-configmap-och-longhorn-node-cr-ownership-reglerna-som-gor-det-saker]]

