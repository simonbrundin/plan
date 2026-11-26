---
title: SOPS age-nyckelroterad; privat nyckel i 1Password item SOPS
type: note
tags: [security, sops, 1password, projekt:infrastructure]
created: 2026-11-26
updated: 2026-11-26
status: active
---

SOPS-nyckeln roterades 2026-11-26 (commit 423e78a9 + docs(secrets) 9e25a79f i infrastructure). Ny recipient: age1e6cmv0vp3ymmhnqv7rrgy2a7wpnnutawqy509wqm3cn7467cps6qupp354. FÄLLA: sops-decrypt-skillen pekar på op://Private/sops-age/key som INTE finns - den privata nyckeln ligger i 1Password item 'SOPS' (Private-valvet), fält 'SOPS Private Key', referens op://Private/SOPS/SOPS Private Key. Rotationen gjordes parallellt i två sessioner och gav non-fast-forward vid push; lösningen var cherry-pick av icke-överlappande commits ovanpå origin/main + separerad docs-commit, eftersom klartexterna var identiska (samma nyckel). Verifiering före push: kustomize build + sops -d på alla .enc.yaml. old-files/authentik/tmp.enc.yaml är en pre-existing trasig fil (inget sops-metadata) från b9d72ca9 - orörda secrets i klartext finns ev. där, bör rensas separat.

## Uppföljning 2026-11-26 (eftermiddag)

Rotationen spreds till alla projekt-repon (dashboard, displayscreen, foodflow,
hitster, homelab, plan, pappa, mailmaster, travelagent). Två fällor noterade:

1. Remote:s auto-commits ("security: rotate SOPS recipient...") re-krypterade
   .enc.yaml men GLÖMDE ofta .sops.yaml och dokumentreferenser - lokala
   commits rebase:ades ovanpå och levererade bara saknade delar.
2. Glömda filer hittades med sweep av gamla recipienten: displayscreens
   ghcr-secret.enc.yaml (endast krypterad till gamla => klustret kunde inte
   dekryptera), mailmasters kameo.env.enc samt hårdkodad default-recipient
   i sops.ts i mailmaster och pappa. Sök mottagare som missats med:
   `git grep age1mv3yrgdtqddlr5v9tg67nydvztm3sm6r3uu3j9zhdp736p0wt5gqwyj9lt`

Endast arkiverade agents-repot har kvar gamla referenser (ändras inte).
