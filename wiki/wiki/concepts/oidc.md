---
title: OIDC
type: concept
summary: OIDC är protokollet som låter Zitadel-autentiserade användare få åtkomst till Kubernetes API-servern.
tags: [oidc, oauth2, identity-provider, authentication, rbac, kubernetes]
sources: [talos/patches/controlplane.yaml, infrastructure-flux/components/auth/, infrastructure-flux/components/zitadel/]
created: 2026-10-23
updated: 2026-10-23
---

# OIDC

## Definition

OIDC (OpenID Connect) är identitetslagret ovanpå OAuth 2.0. I det här
klustret är kedjan:

```
Användare → Zitadel (issuer) → JWT-token → API-server → RBAC (oidc-*-roller)
```

En användare loggar in i Zitadel-portalen på `auth.simonbrundin.com`, får
ett JWT, och använder det som `Authorization: Bearer <token>` mot
Kubernetes API-servern. API-servern validerar tokenet mot Zitadels
issuer, och RBAC styr vad rollen ger.

## Issuer-konfiguration

I `talos/patches/controlplane.yaml` (Talos v1.14):

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

Observera: i Talos v1.14 ligger OIDC-konfigurationen i
`KubeAuthenticationConfig`, inte i `KubeAPIServerConfig.extraArgs` - de
gamla `--oidc-*`-flaggorna avvisas med "use KubeAuthenticationConfig".

## Rollmappning

Zitadel-roller översätts till Kubernetes-ClusterRoles via
`auth.zitadel.org/role`-annotationen:

| Zitadel-roll | ClusterRole | Namespace | Rättigheter |
|--------------|-------------|-----------|-------------|
| viewer | `oidc-viewer` | kluster | Läsning av pods, services, deployments, logs |
| developer | `oidc-developer` | kluster | Utvecklaråtkomst |
| cluster-admin | `oidc-cluster-admin` | kluster | Full åtkomst |

För att komma åt ett specifikt namespace binds rollen med en
`ClusterRoleBinding` (se `clusterrolebinding-template.yaml`).

## Andra OIDC-användare

Utöver API-servern används OIDC även för:

- **Flux UI** - `SecurityPolicy` `flux-ui-oidc` i `flux-system` med
  issuer `https://auth.simonbrundin.com`, callback
  `https://flux.simonbrundin.com/oauth2/callback`.
- **Applikationer** - t.ex. OpenClaw har en `oidc-secret.enc.yaml` och
  egen security policy.

## Kopplingar

- [[concepts/zitadel]]
- [[concepts/authentik]]
- [[concepts/envoy-gateway]]
- [[entities/kubernetes]]
