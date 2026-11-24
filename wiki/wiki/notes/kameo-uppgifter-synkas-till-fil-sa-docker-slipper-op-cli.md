---
title: Kameo-uppgifter synkas till fil så Docker slipper op-cli
type: note
tags: [kameo, 1password, docker, nuxt, projekt:mailmaster]
created: 2026-10-29
updated: 2026-10-29
status: active
---

Lösning på felet 'Executable not found in PATH: op' i containern: 1Password-appen finns bara på värden, så uppgifterna synkas dit med 'bun run kameo:sync' (src/nuxt/scripts/kameo-sync.ts) som skriver environments/dev/kameo.env (600, gitignorerad) med KAMEO_EMAIL, KAMEO_PASSWORD och KAMEO_TOTP_SECRET. docker-compose läser filen via env_file (required: false), så containern får KAMEO_*-variablerna utan op.

Varför: op-cli kräver 1Password-appen och en levande session, vilket inte fungerar i containern eller i en daemoniserad process. TOTP-hemligheten i filen används för att generera koden lokalt, så op anropas max en gång per inloggning.

Uppslagningsordning i server/utils/kameo.ts: miljövariabler -> synkad fil -> op-cli. Felet som visas ska alltid peka ut på 'bun run kameo:sync'.
