---
title: STK Matchi-synk: självåterställande synk och var DATABASE_URL sätts
type: note
tags: [matchi, stk, beslut, gitops, projekt:tmp]
created: 2026-11-06
updated: 2026-11-06
status: active
related: [notes/synk-hallbokningar-kommunens-actor-smartbook-matchi]
---

Synkroniseringen mellan Sundsvalls kommuns bokningssystem (Actor Smartbook) och MATCHi är självåterställande: om en bokning flyttats eller tagits bort direkt i MATCHi flyttas den tillbaka eller bokas om. Det gör synkroniseringen robust mot manuella ingrepp i MATCHi.

Identifiering sker via kommun_appointment_id + matchi_court_id (inte slot-id), så flera kommunresurser kan peka på samma MATCHi-bana.

DATABASE_URL sätts per miljö, inte i base: dev patchar go-api + matchi-sync mot repo-template-db-neon/DATABASE_URL, prod mot repo-template-db-app/uri. Base innehåller bara go-api-secrets (MATCHi-inloggning), krypterat med SOPS + age.

Flux-kustomize-controller måste ha spec.decryption.provider: sops, annars appliceras de krypterade värdena. Detta är inte verifierat ännu - inget Flux-kustomization för repot finns ännu.

## Relaterade

- [[notes/synk-hallbokningar-kommunens-actor-smartbook-matchi]]

