---
title: OIDC-fix: jwt-konfiguration applicerad pa controlplanes och Zitadel-RBAC applicerat
type: note
tags: [oidc, auth, zitadel, cluster1, incident, projekt:simon]
created: 2026-11-25
updated: 2026-11-25
status: active
related: [notes/falla-fixad-talosctl-kubeconfig-skapar-inga-dubbletter-langre-talos-kubeconfig-shell-guard, notes/kubeconfig-stadning-en-kontext-mot-cluster1-admin-cluster1-6-dubbletter-borttagna]
---

## Symptom

`kubectl --context oidc-cluster1` gav "You must be logged in to the server (Unauthorized)"
trots en korrekt Zitadel-token. Admin-kontexten fungerade samtidigt.

## Rotorsak

Tre oberoende blockerare, alla maste fixas for att OIDC ska fungera:

1. **API-serverns JWT-lista var tom.** `talosctl read
   /system/config/kubernetes/kube-apiserver/authentication-config.yaml` visade
   `jwt: []` pa controlplane-2 och -3. Tomma `jwt: []` betyder att apiservern
   bara accepterar client-certifikat - en fullt korrekt OIDC-token ger da 401.
   Repot (talos/patches/controlplane.yaml) hade korrekt Zitadel-config, men
   den var aldrig applicerad pa de korande noderna.

2. **RBAC-bindningarna saknades i klustret.** `zitadel-cluster-admin` och
   `zitadel-viewers` fanns i infrastructure-flux/components/auth/ men inte i
   klustret. Flux `flux-system` hade failat i 123 omgangar (p.g.a.
   `CapabilityScanConfig/dot-ai/default-scan` - Kyverno-CRD saknas), sa
   auth-komponenten applicerades aldrig.

3. **controlplane-1 (10.10.10.11) var nere** ("no route to host") och kunde
   inte patchas. etcd-kvorum fanns tack vare cp-2 och cp-3.

## Fix

- Patchad cp-2 och cp-3 med `talosctl patch machineconfig --patch
  @talos/patches/controlplane.yaml --mode=no-reboot`, EN nod i taget med
  verifiering (etcd-kvorum kravde 2 av 3).
- Applicerade auth-komponenten manuellt via kubectl (Flux kunde inte synka),
  och satte Flux-labels pa de manuella CRB:erna sa att en lyckad sync
  adopterar dem istallet for att prune:a dem.
- Patchen ar applicerad pa cp-1 sa snart noden startar - detta maste goras
  manuellt, den ar inte en Flux-manifest.

## Viktig mapping-detalj

Patchen mappar `sub`-claimen med prefix `zitadel-`, sa tokenens
`sub = 346062090475757945` ger Kubernetes-anvandarnamnet
`zitadel-346062090475757945`. CRB-subject i repot ar samma namn - INTE
email-adressen. README:s felsokningsavsnitt sagde tidigare "kontrollera att
username-claim ar email" - det var fel och gav 401 trots korrekt token.

Audience `391307619412774131` maste matcha tokenens `aud`.

## Kund-kommando

kubectl anvander `kubectl-oidc_login` via exec-plugin, inte kubectl:s
inbyggda `--auth-provider=oidc` (deprecated, saknar refresh och fanns inte i
det faktiska ~/.kube/config).

## Generatorfella (viktig)

`pxe/generate-node-configs.py` bygger controlplane-configer ur en korande
controlplane (`CONTROLPLANE_SOURCE`). KALLANS machine-config bestammer vilken
KubeAuthenticationConfig de genererade noderna far. Med kallan pa en nod med
`jwt: []` hade nya controlplanes bootat UTAN OIDC. Generatorn har nu en
KANDIDATLISTA och hoppar over nere noder (kallan var pa cp-1 som var nere, sa
generatorn gick inte att kora alls). Ordningen ska peka pa en nod MED OIDC.

## Relaterade

- [[notes/falla-fixad-talosctl-kubeconfig-skapar-inga-dubbletter-langre-talos-kubeconfig-shell-guard]]
- [[notes/kubeconfig-stadning-en-kontext-mot-cluster1-admin-cluster1-6-dubbletter-borttagna]]
