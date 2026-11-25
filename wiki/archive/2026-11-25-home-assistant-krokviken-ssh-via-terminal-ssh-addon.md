---
title: Home Assistant Krokviken: SSH via Terminal & SSH-addon
type: note
tags: [home-assistant, ssh, supervisor, projekt:simon]
---

HAOS 18.4 'Krokviken' på 10.20.20.20 (aarch64). Port 22 var 'connection refused' eftersom Terminal & SSH-addonet (core_ssh 10.6.0) hade nätverksmappningen '22/tcp': null (inget värdport publicerat) och loggen säger 'SSH port is disabled'.

Nyckelinsikt: HA cores egna REST-endpoints (/api/hassio/addons, /addons/{slug}/config m.fl.) returnerar 401 även för admin - i HA core dev finns bara PATHS_ADMIN (logs/backups/changelog) och PATHS_NO_AUTH (logo/icon) som tillåts via /api/hassio/*. Administrator-API:t mot supervisor nås i stället via WebSocket-kommandot type='supervisor/api' med endpoint/method/data (kräver admin-token).

Fix som gjordes: via supervisor/api POST /addons/core_ssh/options med {options: {authorized_keys:[<rätta nycklar>, password:'', apks:[], server:{tcp_forwarding:false}}, network: {'22/tcp': 22}} och sedan POST /addons/core_ssh/restart. Den tidigare authorized_keys-raden var dessutom trasig (mellanslag in i nyckeln) och ersattes.

Nu: ssh root@10.20.20.20 fungerar med både ~/.ssh/id_ed25519 och ~/.ssh/stuganha; alias 'krokviken' finns i ~/.ssh/config. Långlivstoken för HA finns i 1Password (item 'Home Assistant').
