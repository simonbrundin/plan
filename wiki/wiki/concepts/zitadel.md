---
title: Zitadel
type: concept
summary: Zitadel är klustrets self-hosted identity provider och utfärdar OIDC-tokens som API-servern accepterar.
tags: [zitadel, identity-provider, oidc, sso, authentication, rbac]
sources: [infrastructure-flux/components/zitadel/, infrastructure-flux/components/auth/, talos/patches/controlplane.yaml]
created: 2026-10-23
updated: 2026-11-26
related: [concepts/oidc, notes/zitadel-aterstallning-historisk-installation-aterstalld-fran-zitadel-cloud-export, notes/zitadel-import-application-id-bevaras-men-oidc-client-id-genereras-om, notes/zitadel-importerad-anvandare-kravde-init-code-event-for-att-kunna-logga-in]
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
        url: https://auth.simonbrundin.com
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

Issuer pekar på den **self-hosade** instansen (`auth.simonbrundin.com`).
Tidigare pekade den på ZITADEL Cloud (`simonbrundin-hvukrq.us1.zitadel.cloud`);
bytet skedde 2026-11-26 i samband med att den historiska installationen
återställdes self-hosted (se [[notes/zitadel-aterstallning-historisk-installation-aterstalld-fran-zitadel-cloud-export]]).

Kontrollera att patchen faktiskt är applicerad på **alla** controlplanes -
en av dem saknade OIDC-konfigurationen (`jwt: []`) innan patchen kördes:

```bash
talosctl patch machineconfig --nodes 10.10.10.11,10.10.10.12,10.10.10.13 \
  --patch @talos/patches/controlplane.yaml --dry-run   # visar diffen per nod
```

## Återställd installation (2026-11-26)

Den historiska installationen (organisation `simon-brundin-ab`,
`346062090475233657`) återställdes från en ZITADEL Cloud-export med
ZITADEL:s eget export/import-flöde. Applikationernas historiska ID:n
behölls, men OIDC `client_id` genererades om vid importen och skrevs
därför tillbaka i eventstore och projection.

| App | Application ID | Client ID | Typ | Används av |
|-----|----------------|-----------|-----|-----------|
| Console | `346062090475561337` | `346062091180007801` | public | `auth.simonbrundin.com/ui/console` |
| talos-oidc | `346062485578709720` | `346062485595421400` | public | Talos OIDC |
| kubernetes-kubectl | `391307619395996915` | `391307619412774131` | public | `kubectl oidc-login` (API-serverns audience) |
| Foodflow | `392291832727364012` | `392291832744141228` | public | Foodflow |
| Plan | `400928499969314176` | `391431970426454645` | public | Plan (`/auth/login`) |
| cluster-ui | `401364362276635673` | `392027778305753717` | confidential | Flux, Dashboard, Longhorn, OpenClaw (via Envoy SecurityPolicy) |

Användaren `simonbrundin@gmail.com` (`346062090475757945`) är ORG_OWNER i
organisationen och har samma user-ID i både Cloud- och self-hosted-instansen,
vilket gör att Kubernetes-username (`zitadel-346062090475757945`) är stabil
oavsett vilken issuer API-servern använder.

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

## Återställd installation

Den historiska installationen (ZITADEL Cloud) är återställd i den
self-hostade instansen via ZITADEL:s officiella export/import. Importen skapar
en **ny instans** (`401310226394907493`) - den gamla instansens PAT och
service account-nycklar gäller alltså inte de importerade resurserna, och
management-API:t svarar `User could not be found` för dem.

- Organisation `346062090475233657` (simon-brundin-ab), projekt `talos-oidc`,
  `kubernetes-oidc`, `Foodflow`, `Plan`.
- Användaren `simonbrundin@gmail.com` (`346062090475757945`).
- `Application ID` bevaras vid import, men OIDC `client_id` genereras om -
  se [[notes/zitadel-import-application-id-bevaras-men-oidc-client-id-genereras-om]].
- Administratörens inloggning ligger krypterat i `zitadel-user-credentials`
  (SOPS) i infrastructure-repot.

## Relaterade

- [[notes/zitadel-importerade-anvandarens-losenordshash-stammde-inte-med-nagot-kant-losenord]]
- [[notes/zitadel-importerad-anvandare-kravde-init-code-event-for-att-kunna-logga-in]]
- [[notes/zitadel-import-application-id-bevaras-men-oidc-client-id-genereras-om]]
- [[notes/zitadel-aterstallning-historisk-installation-aterstalld-fran-zitadel-cloud-export]]
- [[concepts/oidc]]
