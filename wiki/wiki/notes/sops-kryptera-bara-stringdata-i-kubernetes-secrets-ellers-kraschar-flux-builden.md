---
title: SOPS - kryptera bara stringData i Kubernetes-secrets, annars kraschar Flux-builden
type: note
summary: Med SOPS' standard encrypted_regex krypteras även metadata, och Flux' SOPS-decryption panikar i kustomize-builden med "number of previous names, number of previous namespaces, number of previous kinds not equal".
tags: [sops, secrets, flux, bug, projekt:plan]
created: 2026-11-26
updated: 2026-11-26
status: active
related: [concepts/sops, concepts/flux-operator]
---

# SOPS - kryptera bara stringData i Kubernetes-secrets, annars kraschar Flux-builden

Efter att `plan-db-user` ändrades (se
[[notes/cnpg-managed-roles-secretet-saknar-losenord-om-username-avviker-fran-rollnamnet]])
 började Flux-fsynkningen av Plan misslyckas utan ett tydligt fel:

```
number of previous names, number of previous namespaces, number of previous kinds not equal
```

Det kom från kustomize/Flux inuti byggsteget, inte från API-servern.

## Varför

SOPS' `encrypted_regex` styr vilka fält som krypteras. Utan regex krypteras
**alla** strängar - inklusive `metadata.name`, `metadata.namespace` och
`metadata.labels`. När Flux dekrypterar sekretet och bygger vidare med
kustomize jämför den en krypterad `metadata` mot den dekrypterade originalet,
och jämförelsen faller isär - "number of previous names ... not equal".

För Kubernetes Secrets ska krypteringen begränsas till payloaden:

```yaml
sops:
  encrypted_regex: ^(data|stringData)$
  age:
    - recipient: age1e6cmv0vp3ymmhnqv7rrgy2a7wpnnutawqy509wqm3cn7467cps6qupp354
  version: 3.13.3
```

## Lösning

Lägg till `encrypted_regex: ^(data|stringData)$` i varje `*.enc.yaml` som
innehåller ett Kubernetes Secret, och kryptera om filen:

```bash
sops --encrypt --in-place secrets.enc.yaml
```

Verifiera att bygget blir rent igen:

```bash
kustomize build environments/base > /dev/null
```

## Kopplingar

- [[concepts/sops]]
- [[concepts/flux-operator]]
