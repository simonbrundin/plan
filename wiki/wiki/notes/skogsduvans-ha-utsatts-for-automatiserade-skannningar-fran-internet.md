---
title: Skogsduvans HA utsätts för automatiserade skannningar från internet
type: note
tags: [home-assistant, skogsduvan, säkerhet, session, projekt:simon]
created: 2026-11-03
updated: 2026-11-03
status: active
related: [notes/skogsduvans-ha-temperaturlogik-ombyggd-fyra-doda-sensorer]
---

2026-11-03: I Home Assistants logg på Skogsduvan syntes WARNINGar från homeassistant.components.http.security_filter om filtrerade förfrågningar med "/index.php?lang=../../../../../../../../usr/local/lib/php/pearcmd&+config-create+/&/<?echo(md5(\"hi\"));?>+/tmp/index1.php" och "/index.php?lang=../../../../../../../../tmp/index1".

Det är automatiska skannningar efter PHP-körbara filer - en bilaga till HA men inte en kompromissering. security_filter blockade dem. Varningarna är alltså benigna, men de visar att HA:s webbgränssnitt NÅS FRÅN INTERNET (troligen via den duckdns-tunnel/port-forward som används för SSH).

Att tänka på: (1) inga personuppgifter eller credentials i sådana noter - bara mönstret, (2) HA:s egen brandvägg (security_filter) fångar detta men är inte en ersättning för åtkomstkontroll - overväg att sla HA bakom en autentiserad tunnel (Cloudflare Access, Tailscale) eller en VPN, (3) håll koll pa att inga HA-token eller loggar med hemligheter hamnat i wikin.

## Relaterade

- [[notes/skogsduvans-ha-temperaturlogik-ombyggd-fyra-doda-sensorer]]

