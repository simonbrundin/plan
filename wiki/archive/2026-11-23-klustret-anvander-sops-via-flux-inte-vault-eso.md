---
title: Klustret anvander SOPS via Flux - inte Vault/ESO
type: note
tags: [secrets, gitops, sops, projekt:infrastructure]
---

Klustret (Talos, home) har EN secret-vag: SOPS + age, dekrypterat av Flux kustomize-controller via spec.decryption.provider: sops (patch i infrastructure-flux/components/flux/flux-instance.yaml, hemlighet: Secret sops-age i flux-system). Det finns INGEN Vault och INGEN External Secrets Operator.

Varfor detta behovs: flera repon och AGENTS.md-filer har fortfarande beskrivet en ESO/Vault-modell (t.ex. plan/AGENTS.md: 'Vault is down (no pods in vault namespace) - ESO cant sync secrets / Auth secret patched manually'). Det beskriver en arbetsmodell som inte finns - i klustret finns varken vault-namespace eller ESO-CRDs (SecretStore/ClusterSecretStore/ExternalSecret), och det finns ingen vault-komponent i infrastructure-flux/components/. ESO/Vault-materialet ligger kvar i infrastructure/old-files/argo-manifests/ (ArgoCD-era, arkiverat) - det ar historik, inte nutid.

Verifiering (2026-11-23): kubectl get crd visar inga ESO-CRDs; kubectl get ns -n vault ger 'No resources found'; grep -ri vault|external-secret i infrastructure-flux/ trarffar bara Longhorn-storageclassen 'longhorn-vault'. Public age-nyckel: infrastructure-flux/components/secrets/sops.age.pub (age1mv3yrgdtqddlr5v9tg67nydvztm3sm6r3uu3j9zhdp736p0wt5gqwyj9lt) - SAME key anvands i alla app-repon, t.ex. plan/environments/base/sops-public-key.yaml. Privat nyckel: ~/.config/sops-age/keys.txt (aldrig commit).

FReger framot: nya hemligheter = *.enc.yaml (SOPS + age) i app-repots environments/<env>/secrets/, appliceras av Flux. Devi miljoer utan Flux: dekryptera lokalt till en gitignoras .env och anvanda kustomize secretGenerator (mönstret fran db/neon.env). Skapa ALDRIG en 'manuellt patchad secret' i klustret - det blir drift utan kalla, och en ENC[...]-rad som appliceras rakt av ar en inloggningsattack mot egen stack (se noten om Keycloak-brute force).
