---
title: Lunchdeploy triggas från cronjob på pxe
type: note
tags: [ci, lunch, github-actions, pxe, projekt:stk]
created: 2026-11-23
updated: 2026-11-23
status: active
---

Veckovis deploy av lunchsidan triggas från **pxe-serven** (Raspberry Pi,
Ubuntu, tidszon Europe/Stockholm) via en crontab-entry:
`0 3 * * 1 /home/simon/bin/trigger-lunch-deploy.sh`.

Varför inte GitHub Actions: GHA-schemaläggning är UTC-baserad och kan
fördröjas, så 03:00 blir 02:00/04:00 UTC beroende på sommar-/vintertid.
En cronjob på pxe ger exakt 03:00 lokal tid året om.

Skriptet (versionerat i stk-repot som `scripts/trigger-lunch-deploy.sh`)
POST:ar till `https://api.github.com/repos/simonbrundin/stk/actions/workflows/pages.yaml/dispatches`
med `ref=main`. Token är en PAT med "Actions: read and write" på stk,
lagd på pxe i `~/.config/stk/github-token` (0600). Logg: `~/log/lunch-deploy.log`
(HTTP 204 = ok).

Lärdom: en finstark PAT som bara har repo-behörigheter räcker INTE för att
dispatch:a workflows ("Resource not accessible by personal access token").
En PAT som används för push-taggar (klassisk PAT) fungerade direkt.

Övrigt: `pages.yaml` har fortfarande en daglig schemaläggning kl 06:00
svensk tid, så veckotrigger + dagsbygge överlappar.
