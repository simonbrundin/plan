---
title: SOPS age-nyckelroterad; privat nyckel i 1Password item SOPS
type: note
tags: [security, sops, 1password, projekt:infrastructure]
created: 2026-11-26
updated: 2026-11-26
status: active
---

SOPS-nyckeln roterades 2026-11-26 (commit 423e78a9 + docs(secrets) 9e25a79f i infrastructure). Ny recipient: age1e6cmv0vp3ymmhnqv7rrgy2a7wpnnutawqy509wqm3cn7467cps6qupp354. FÄLLA: sops-decrypt-skillen pekar på op://Private/sops-age/key som INTE finns - den privata nyckeln ligger i 1Password item 'SOPS' (Private-valvet), fält 'SOPS Private Key', referens op://Private/SOPS/SOPS Private Key. Rotationen gjordes parallellt i två sessioner och gav non-fast-forward vid push; lösningen var cherry-pick av icke-överlappande commits ovanpå origin/main + separerad docs-commit, eftersom klartexterna var identiska (samma nyckel). Verifiering före push: kustomize build + sops -d på alla .enc.yaml. old-files/authentik/tmp.enc.yaml är en pre-existing trasig fil (inget sops-metadata) från b9d72ca9 - orörda secrets i klartext finns ev. där, bör rensas separat.
