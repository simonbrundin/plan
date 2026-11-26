---
title: SOPS
type: concept
summary: SOPS krypterar secrets i git med age - Flux dekrypterar dem via sops-age och provider: sops.
tags: [secrets, encryption, age, kubernetes, gitops]
sources: [infrastructure-flux/components/secrets/sops.age.pub, infrastructure-flux/components/flux/flux-instance.yaml]
created: 2026-06-02
updated: 2026-11-23
related: [entities/kubernetes, entities/talos-linux, notes/klustret-anvander-sops-via-flux-inte-vault-eso]
---

# SOPS

## Definition

SOPS (Secrets OPerationS) är ett verktyg för att kryptera/decryptera
hemligheter (YAML, JSON, ENV) med age-nycklar eller PGP. I det här
repot används **age**.

## I klustret

Hemligheter i `infrastructure-flux/components/secrets/` är krypterade med
SOPS. Den publika mottagarnyckeln ligger i
`infrastructure-flux/components/secrets/sops.age.pub`:

```
age1e6cmv0vp3ymmhnqv7rrgy2a7wpnnutawqy509wqm3cn7467cps6qupp354
```

Flux dekrypterar automatiskt via patchen i `flux-instance.yaml`:

```yaml
  kustomize:
    patches:
      - patch: |
          - op: add
            path: /spec/decryption
            value:
              provider: sops
              secretRef:
                name: sops-age
        target:
          kind: Kustomization
```

Utan den patchen bryts varje krypterat secret i klustret.

## Ingen Vault, ingen External Secrets Operator

SOPS + Flux är klustrets **enda** secret-väg. Det finns ingen Vault och
ingen ESO (External Secrets Operator) i klustret:

- Inga ESO-CRDs (`SecretStore`, `ClusterSecretStore`, `ExternalSecret`).
- Inget `vault`-namespace och ingen vault-komponent i
  `infrastructure-flux/components/`.
- ESO/Vault-materialet som finns ligger i
  `infrastructure/old-files/argo-manifests/` - det är ArgoCD-era och
  arkiverat, inte nutid.

Om en repo-`AGENTS.md` eller ett felmeddelande nämner "Vault is down",
"ESO kan inte synka secrets" eller "manuellt patchat secret", är det
**stale information** - följ inte den vägen. Lösningen är att kryptera
med SOPS och låta Flux applicera.

Verifiering (2026-11-23):

```bash
kubectl get crd | grep -iE "secretstore|externalsecret"   # tomt
kubectl get ns vault                                      # No resources found
grep -ri "vault\|external-secret" ~/repos/infrastructure/infrastructure-flux/
```

## Nycklar och repos

Samma age-nyckel används i hela "estate":

| Plats | Innehåll |
|-------|----------|
| `infrastructure-flux/components/secrets/sops.age.pub` | Publik nyckel (canonical) |
| `infrastructure-flux/components/secrets/sops-age-secret.enc.yaml` | Privat nyckel, krypterad, appliceras av Flux |
| `~/.config/sops-age/keys.txt` | Privat nyckel lokalt (aldrig commit) |
| `<app>/environments/base/sops-public-key.yaml` | Samma publika nyckel i app-repon |

## Lägga till en ny hemlighet

```bash
sops --encrypt --age age1e6cmv0vp3ymmhnqv7rrgy2a7wpnnutawqy509wqm3cn7467cps6qupp354 \
  secrets.yaml > secrets.enc.yaml
git add -f secrets.enc.yaml     # *.enc.yaml är gitignorerat
```

I miljöer utan Flux (dev) dekrypteras lokalt till en gitignoras `.env`
och matas in via kustomizes `secretGenerator` - samma mönster som
`db/neon.env`. Applicera **aldrig** en `ENC[...]`-rad rakt av som
credential; det har redan orsakat ogiltiga inloggningar mot stacken.

## Filnamnskonvention

Krypterade filer har suffixet `*.enc.yaml`. `.gitignore` matchar
`*.enc.yaml`, så **en ny** krypterad fil måste läggas in med
`git add -f` - annars når den aldrig klustret. De 22 redan spårade
hemligheterna ligger kvar som tracked files.

## Grundläggande användning

```bash
# Kryptera
sops --encrypt --age age1e6cmv0vp3ymmhnqv7rrgy2a7wpnnutawqy509wqm3cn7467cps6qupp354 \
  secrets.yaml > secrets.enc.yaml

# Dekryptera
sops --decrypt secrets.enc.yaml
```

## Age-nycklar

Age är det moderna alternativet till PGP för SOPS. Nycklar genereras
med:

```bash
age-keygen -o key.txt
```

Den privata nyckeln ska aldrig committas.

## Kopplingar

- [[concepts/flux-operator]]
- [[entities/flux-gitops]]
- [[entities/kubernetes]]
- [[entities/talos-linux]]
