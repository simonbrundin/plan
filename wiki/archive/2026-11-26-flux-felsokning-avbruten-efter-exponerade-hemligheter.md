---
title: Flux-felsökning avbruten efter exponerade hemligheter
type: note
tags: [incident, flux, säkerhet, projekt:plan]
---

Vid felsökning av cluster1 exponerades en GitHub-token och en privat SOPS age-nyckel i agentens verktygsutdata.

Åtgärder 2026-11-26:
- Den läckta klassiska GitHub-tokenen återfanns utan att skrivas ut igen, återkallades via GitHubs anonyma credential-revocation-API (HTTP 202) och verifierades därefter ogiltig.
- Flux GitHub-hemligheten byttes till en fungerande, fininställd PAT från 1Password. Klustrets Secret och 1Password-värdet verifierades lika.
- En ny age-identitet skapades och sparades i 1Password. Den distribuerades till Flux och ersatte den gamla identiteten; den lokala `~/.config/sops-age/keys.txt` byttes också.
- 50 aktuella SOPS-filer i nio repos skrevs om så att endast den nya mottagarnyckeln kan dekryptera dem. Publika nyckelreferenser uppdaterades. Filerna verifierades med nya identiteten.
- Ändringarna pushades till respektive `main`; de två skyddade brancherna `infrastructure` och `plan` krävde bypass av PR-regeln.
- `flux-system` och `bootstrap` är nu Ready på infrastrukturrevision `423e78a9`. `flux-operator` är Ready. De två deploymenterna som felaktigt togs bort under incidenten är återställda.

Kvarvarande säkerhetsrisk: omskrivning av nuvarande SOPS-filer raderar inte Git-historik. Eftersom den gamla privata nyckeln exponerades kan äldre SOPS-versioner i historiken fortfarande dekrypteras. De underliggande credentials som fanns i historiska snapshots bör därför inventeras och roteras/revokeras. Displayscreens `ghcr-secret.enc.yaml` dekrypterades inte med den komprometterade identiteten och lämnades oförändrad; den filen använder en annan mottagare.

Kvarvarande driftfel är separata från nyckelrotationen: Longhorn-webhook-timeout och workload-problem kan fortfarande blockera apparnas reconciliation.
