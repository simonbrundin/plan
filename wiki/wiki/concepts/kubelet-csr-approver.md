---
title: Kubelet CSR Approver
type: concept
summary: Kubelet CSR Approver är en Kubernetes-controller som automatiskt godkänner Certificate Signing Requests (CSRs) från kubelet-aracters base...
tags: [kubelet, csr, certificate, pki, node-authentication]
sources: [infrastructure-flux/components/kubelet-csr-approver/]
created: 2026-06-02
updated: 2026-10-24
---

# Kubelet CSR Approver

## Definition

Kubelet CSR Approver är en Kubernetes-controller som automatiskt godkänner Certificate Signing Requests (CSRs) från kubelet-aracters baserat på konfigurerbara regler. Detta förenklar nodregistrering genom att automatisera godkännandeprocessen.

## Nyckelpunkter

### Installation via Flux

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: kubelet-csr-approver
  namespace: kubelet-csr-approver
spec:
  chart:
    spec:
      chart: kubelet-csr-approver
      version: 1.1.0
      sourceRef:
        kind: HelmRepository
        name: kubelet-csr-approver
  values:
    providerRegex: "^(controlplane-|worker-).*$"
    bypassDnsResolution: false
    leaderElection: true
    replicas: 1
    metrics:
      enable: true
      serviceType: ClusterIP
      port: 8080
    tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/control-plane
        operator: Exists
```

### Hur det fungerar

```
1. Nod bootar med kubelet config
       ↓
2. Kubelet genererar CSR för node certifikat
       ↓
3. kubelet-csr-approver övervakar nya CSRs
       ↓
4. CSR matchar providerRegex (controlplane-|worker-.*)
       ↓
5. CSR godkänns automatiskt
       ↓
6. Kubelet får certifikat och registrerar sig
```

### Viktiga konfigurationsvärden

| Parameter | Standard | Beskrivning |
|-----------|----------|-------------|
| `providerRegex` | `^(controlplane-\|worker-).*$` | Matchar nodnamn |
| `bypassDnsResolution` | `false` | Om DNS-upplösning ska hoppas över |
| `replicas` | `1` | Antal repliker |
| `leaderElection` | `true` | Aktiverar leader election |

### Provider Regex

Regex: `^(controlplane-|worker-).*$`

| Nodnamn | Matchar |
|---------|---------|
| controlplane-1 | ✅ |
| controlplane-2 | ✅ |
| worker-1 | ✅ |
| worker-9 | ✅ |
| my-custom-node | ❌ |

### Certifikat-typer som godkänns

| CSR Type | CN-format |
|----------|----------|
| kubelet-serving | `system:node:<nodename>` |
| kubelet-client | `system:node:<nodename>` |

### Kommandon för felsökning

```bash
# Lista alla CSRs
kubectl get csr

# Visa specifik CSR
kubectl get csr <csr-name> -o yaml

# Beskriv CSR (inkl. events)
kubectl describe csr <csr-name>

# Lista godkända CSRs
kubectl get csr | grep Approved

# Visa approver logs
kubectl logs -n kubelet-csr-approver -l app=kubelet-csr-approver

# Kontrollera om approver körs
kubectl get pods -n kubelet-csr-approver
```

### Vanliga problem

| Problem | Orsak | Lösning |
|---------|-------|---------|
| CSR inte godkänd | Nodnamn matchar inte regex | Verifiera `providerRegex` |
| CSR väntar | Approver inte installerad | Installera kubelet-csr-approver |
| `Unable to decode certificate` | Fel format i CSR | Kontrollera kubelet-konfiguration |
| Approver Pod crashloop | RBAC-problem | Verifiera service account och roller |

### RBAC-krav

Kubelet CSR Approver kräver följande RBAC:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kubelet-csr-approver
rules:
  - apiGroups: ["certificates.k8s.io"]
    resources: ["certificatesigningrequests"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["certificates.k8s.io"]
    resources: ["certificatesigningrequests/approval"]
    verbs: ["create", "update", "patch"]
  - apiGroups: ["certificates.k8s.io"]
    resources: ["signers"]
    verbs: ["approve"]
```

## Kopplingar

- [[entities/kubernetes]]
- [[entities/talos-linux]]
