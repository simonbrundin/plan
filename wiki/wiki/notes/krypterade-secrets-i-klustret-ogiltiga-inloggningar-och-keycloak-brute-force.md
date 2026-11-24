---
title: Krypterade secrets i klustret: ogiltiga inloggningar och Keycloak-brute force
type: note
tags: [bug, sops, matchi, keycloak, lesson, gitops, projekt:tmp]
created: 2026-11-06
updated: 2026-11-06
status: active
related: [notes/synk-hallbokningar-kommunens-actor-smartbook-matchi]
---

Problem: SOPS-krypterade värden applicerades rakt av i dev-klustret (ingen Flux där). Secretet go-api-secrets innehöll då ENC[AES256_GCM,...] som skickades som MATCHi-lösenord. Keycloak räknar det som ogiltiga inloggningsförsök - risk att kontot låses av brute force-skyddet.

Fixar i stk:
1) Klienten vägrar logga in om lösenordet börjar med ENC[ (isEncryptedSecret) - inga försök mot Keycloak alls. Test: TestLoginRejectsEncryptedSecrets.
2) Hemlighetsreferenserna i deploymenten och cronjobbet har optional: true, så podden startar även i miljöer utan secret (t.ex. PR-preview).
3) Det krypterade secreten ligger i environments/prod/secrets/ (Flux dekrypterar med spec.decryption.provider: sops). Dev dekrypterar lokalt via environments/dev/secrets/sync-matchi-env.sh till en gitignoras matchi.env som kustomizes secretGenerator använder - samma mönster som db/neon.env.

Viktig regel: krypterade hemligheter får aldrig appliceras som credentials. Kontrollera alltid vad som hamnar i klustret (kubectl get secret -o jsonpath).

Bonusfynd: Tilt ser inte ändringar i ../base eftersom base ligger utanför kustomize-roten - watch_file('../../base/...') i Tiltfilen behövdes för att ändringar i base skulle appliceras.

## Relaterade

- [[notes/synk-hallbokningar-kommunens-actor-smartbook-matchi]]

