---
title: Zitadel
type: concept
summary: Zitadel är klustrets self-hosted identity provider och utfärdar OIDC-tokens som API-servern accepterar.
tags: [zitadel, identity-provider, oidc, sso, authentication, rbac]
sources: [infrastructure-flux/components/zitadel/, infrastructure-flux/components/auth/, talos/patches/controlplane.yaml]
created: 2026-10-23
updated: 2026-10-23
related: [concepts/oidc]
---

# Zitadel

## Definition

Zitadel är klustrets identity provider. Den är self-hosted (HelmRelease,
chart `zitadel` 10.0.2) och ligger i namespace `zitadel`. Zitadel utfärdar
OIDC-tokens som Kubernetes API-servern accepterar, och dess roller styr
vilken RBAC-roll en användare får.

Zitadel ersatte Authentik - se [[concepts/authentik]].

## Komponenter

| Komponent | Namespace | Ansvar |
|-----------|-----------|--------|
| zitadel | zitadel | Identity provider, portal på `auth.simonbrundin.com` |
| zitadel-db | zitadel | CloudNativePG-cluster (1 instans, 10Gi, storageClass `cnpg`) |
| auth | auth | ClusterRoles som mappar Zitadel-roller till Kubernetes-roller |

### Resurser

| Fil                        | Typ            | Beskrivning                     |
|----------------------------|----------------|---------------------------------|
| namespace.yaml             | Namespace      | zitadel namespace               |
| source.yaml                | HelmRepository | `https://charts.zitadel.com`    |
| release.yaml               | HelmRelease    | chart 10.0.2                    |
| db.yaml                    | Cluster (CNPG) | Databas för Zitadel             |
| httproute.yaml             | HTTPRoute      | Gateway API-route till portalen |
| zitadel-config.enc.yaml    | Secret (SOPS)  | Zitadel-konfiguration           |
| masterkey-secret.enc.yaml  | Secret (SOPS)  | Master key                      |
| admin-password.enc.yaml    | Secret (SOPS)  | Lösenord för första admin       |
| db-secret.enc.yaml         | Secret (SOPS)  | Databas-DSN                     |

## OIDC mot Kubernetes API-servern

API-servern konfigureras via `KubeAuthenticationConfig` i
`talos/patches/controlplane.yaml`:

```yaml
apiVersion: v1alpha1
kind: KubeAuthenticationConfig
configuration:
  jwt:
    - issuer:
        url: https://simonbrundin-hvukrq.us1.zitadel.cloud
        audiences:
          - "391307619412774131"
      claimMappings:
        username:
          claim: sub
          prefix: "zitadel-"
```

Viktigt: i Talos v1.14 ligger OIDC-konfigurationen i detta dokument,
inte i `KubeAPIServerConfig.extraArgs` - flaggorna avvisas med "use
KubeAuthenticationConfig".

## RBAC-mappning

Zitadel-roller mappas till Kubernetes via ClusterRoles i
`infrastructure-flux/components/auth/`:

| ClusterRole | Zitadel-annotation | Rättigheter |
|-------------|--------------------|-------------|
| `oidc-viewer` | `auth.zitadel.org/role: viewer` | Läsning av resurser, pod-loggar |
| `oidc-developer` | `auth.zitadel.org/role: developer` | Utvecklaråtkomst |
| `oidc-cluster-admin` | `auth.zitadel.org/role: cluster-admin` | Full klusteråtkomst |

En användare får sin roll genom `groups`-claimet, och
`username`-claimet mappas från `sub` med prefixet `zitadel-`.

## Route

`zitadel-route` är en Gateway-API-`HTTPRoute` som pekar på
`simons-gateway` (namespace `envoy`, sektion `https`) med hostname
`auth.simonbrundin.com`:

- `/ui/v2/login` → `zitadel-login` port 3000
- `/` → `zitadel` port 8080

## Felsökning

```bash
# Pod-status
kubectl get pods -n zitadel

# Databas
kubectl get cluster -n zitadel

# Route
kubectl get httproute -n zitadel
```

## Kopplingar

- [[concepts/authentik]]
- [[concepts/envoy-gateway]]
- [[concepts/cloudnative-pg]]
- [[entities/kubernetes]]

## Relaterade

- [[concepts/oidc]]
