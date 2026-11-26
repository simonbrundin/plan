---
description: >-
  Felsök FluxCD GitOps-operator och föreslå åtgärder. Utlösare: felsök flux, debug flux, flux problem, flux issue, flux troubleshooting.
---

# Debug-flux

## Syfte

Felsöka FluxCD GitOps-installation i Kubernetes-kluster. Kommandot kör
diagnostik, analyserar problem och föreslår konkreta åtgärder för att lösa
identifiererade fel. Kommandot använder `flux-debug`-skillen som referens för
effektiv felsökning.

## Variabler

$KOMMANDO: debug-flux
$ÅTGÄRD: skapa
$REPOSITORY: infrastructure
$FLUX_SKILL: flux-debug

## Workflow

### 1. Hämta Flux-hälsa

Kör följande diagnostik-kommandon i parallell för att få en översikt:

```bash
flux check --verbose
kubectl get all -n flux-system
kubectl get gitrepositories,helmrepositories,buckets,kustomizations,helmreleases -A
```

### 2. Analysera Flux-controllers

Kör diagnostik på varje Flux-controller:

```bash
# Source-controller
kubectl logs -n flux-system -l app=source-controller --tail=100

# Kustomize-controller
kubectl logs -n flux-system -l app=kustomize-controller --tail=100

# Helm-controller
kubectl logs -n flux-system -l app=helm-controller --tail=100

# Notification-controller
kubectl logs -n flux-system -l app=notification-controller --tail=100

# Events för fel/warning
kubectl get events -n flux-system --field-selector type=Warning --sort-by='.lastTimestamp' | tail -30
```

### 3. Analysera problematiska resurser

För varje resurs med problem, kör djupare diagnostik:

```bash
# GitRepository med fel
kubectl describe gitrepository <name> -n <namespace>
kubectl get gitrepository <name> -n <namespace> -o jsonpath='{.status}'

# Kustomization med fel
kubectl describe kustomization <name> -n <namespace>
flux tree kustomization <name> -n <namespace>

# HelmRelease med fel
kubectl describe helmrelease <name> -n <namespace>
kubectl get helmrelease <name> -n <namespace> -o jsonpath='{.status}'
```

### 4. Generera åtgärdsförslag

Baserat på diagnosen, ge specifika åtgärdsförslag:

- **Source-problem**: Verifiera credentials, URL, och branch
- **Dependency-problem**: Använd `flux tree kustomization` för att se beroenden
- **RBAC-problem**: Kontrollera ServiceAccount och ClusterRole-bindings
- **Drift-problem**: Använd `flux diff kustomization` för att se ändringar

## Relevanta filer

- `/home/simon/repos/infrastructure/infrastructure-flux/components/`  
  → Flux-komponentdefinitioner hanterade via Kustomize

- `/home/simon/.pi/agent/skills/flux-debug/SKILL.md`  
  → Flux-felsöknings-skill med detaljerade kommandon och lösningar

## Rapport

Presentera felsökningsresultatet strukturerat:

### Sammanfattning

Kort sammanfattning av Flux-klusterhälsa med antal problem identifierade.

### Identifierade problem

Lista över problem med:

- **Resurs**: namnet på den problematiska resursen
- **Typ**: GitRepository, Kustomization, HelmRelease, etc.
- **Symptom**: vad som är fel (t.ex. "NotReady", "DependencyNotReady")
- **Möjlig orsak**: baserat på loggar och events

### Åtgärdsförslag

För varje problem, ge konkreta åtgärder:

1. **Kommando att köra**: t.ex. `flux reconcile source git <name>`
2. **Filer att granska**: t.ex. `infrastructure-flux/components/<component>/release.yaml`
3. **Verifieringssteg**: hur man bekräftar att problemet är löst

### Återstående steg

Prioriterad lista över åtgärder att genomföra för att återställa Flux.
