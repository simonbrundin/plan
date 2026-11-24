---
title: Cert-Manager
type: concept
summary: cert-manager är en Kubernetes-operator som automatiserar hantering av TLS-certifikat.
tags: [cert-manager, tls, certificates, lets-encrypt, ssl]
sources: [infrastructure-flux/components/cert-manager/]
created: 2026-06-02
updated: 2026-10-24
---

# Cert-Manager

## Definition

cert-manager är en Kubernetes-operator som automatiserar hantering av TLS-certifikat. Den integrerar med Let's Encrypt och andra Certificate Authorities (CA) för att tillhandahålla och förnya certifikat automatiskt.

## Nyckelpunkter

### Installation via Flux

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: cert-manager
  namespace: cert-manager
spec:
  chart:
    spec:
      chart: cert-manager
      version: 1.18.2
      sourceRef:
        kind: HelmRepository
        name: jetstack
  values:
    installCRDs: true
```

### HelmRepository

```yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: jetstack
  namespace: flux-system
spec:
  interval: 1h
  url: https://charts.jetstack.io
```

### Resurser som hanteras

| Fil                                  | Typ            | Beskrivning                |
| ------------------------------------ | -------------- | -------------------------- |
| namespace.yaml                       | Namespace      | cert-manager namespace     |
| source.yaml                          | HelmRepository | Jetstack chart repo        |
| release.yaml                         | HelmRelease    | Helm release               |
| clusterissuer.yaml                   | ClusterIssuer  | Klusterövergripande issuer |
| certificate.yaml                     | Certificate    | Specifika certifikat       |
| cloudflare-api-token-secret.enc.yaml | Secret         | SOPS-krypterad API-token   |
| reference-grant.yaml                 | ReferenceGrant | Gateway API referens       |

### Issuers

**ClusterIssuer** `cloudflare-clusterissuer` för Let's Encrypt:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: cloudflare-clusterissuer
spec:
  acme:
    email: simonbrundin@gmail.com
    server: https://acme-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      name: cloudflare-clusterissuer-account-key
    solvers:
      - dns01:
          cloudflare:
            apiTokenSecretRef:
              name: cloudflare-api-token-secret
              key: api-token
```

Issuern heter `cloudflare-clusterissuer` - det är namnet som Envoy
Gatewayets `Gateway` refererar till i `certificateRefs`.

### Cloudflare-integration

cert-manager använder Cloudflare DNS för DNS-01 challenge:

1. **Challenge:** cert-manager skapar en DNS-txt-post på _acme-challenge.domain.com
2. **Verification:** Let's Encrypt verifierar ägande via DNS
3. **Issuing:** Certifikat utfärdas efter lyckad verification

### Kommandon för felsökning

```bash
# Lista certifikat
kubectl get certificate -A

# Visa certifikat-status
kubectl describe certificate <name> -n <namespace>

# Visa challenge
kubectl get challenge -A

# Visa CertificateRequest
kubectl get certificaterequest -A

# Kontrollera orders
kubectl get order -A
```

### Vanliga problem

| Problem                      | Orsak                         | Lösning                                           |
| ---------------------------- | ----------------------------- | ------------------------------------------------- |
| `Waiting for Certificate...` | DNS challenge inte verifierad | Kontrollera Cloudflare API-token, DNS propagation |
| `error getting certificate`  | ClusterIssuer felkonfigurerad | Verifiera `cloudflare-clusterissuer`            |
| `context deadline exceeded`  | Rate limiting eller timeout   | Vänta och prova igen, kolla Let's Encrypt quotas  |

## Kopplingar

- [[concepts/gitops]]
- [[entities/kubernetes]]
- [[entities/talos-linux]]
