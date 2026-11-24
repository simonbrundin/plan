---
title: Authentik
type: concept
summary: Authentik var tidigare identity provider i klustret och är nu ersatt av Zitadel.
tags: [authentik, identity-provider, oidc, saml, sso, deprecated]
sources: [infrastructure-flux/components/kustomization.yaml]
created: 2026-06-02
updated: 2026-10-23
status: deprecated
related: [concepts/oidc]
---

# Authentik

## Status: deprecated

Authentik **används inte längre**. Identity provider-rollen har gått över
till [[concepts/zitadel]], som är self-hosted och används för OIDC mot
Kubernetes API-servern.

I `infrastructure-flux/components/kustomization.yaml` ligger Authentik
utkommenterad:

```yaml
  # - authentik
  - zitadel
```

## Definition

Authentik är en open-source identity provider som tillhandahåller SSO
(Single Sign-On), OIDC och SAML-autentisering för applikationer.

## Resurser som hanterades

| Fil                    | Typ            | Beskrivning                        |
| ---------------------- | -------------- | ---------------------------------- |
| namespace.yaml         | Namespace      | authentik namespace                |
| source.yaml            | HelmRepository | Authentik chart repo               |
| release.yaml           | HelmRelease    | Helm release                       |
| consent-blueprint.yaml | ConfigMap      | Consent blueprint                  |
| db.yaml                | Cluster        | CNPG Cluster för authentik         |
| secret.yaml            | Secret         | SOPS-krypterad postgres-credential |
| http-route.yaml        | HTTPRoute      | Gateway API route                  |

## Kommandon för felsökning

```bash
# Lista authentik pods
kubectl get pods -n authentik

# Status via Health endpoint
curl -k https://auth.simonbrundin.com/-/health/
```

## Kopplingar

- [[concepts/zitadel]]
- [[concepts/cloudnative-pg]]
- [[entities/kubernetes]]

## Relaterade

- [[concepts/oidc]]
