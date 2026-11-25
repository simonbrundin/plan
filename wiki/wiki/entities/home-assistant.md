---
title: Home Assistant
type: entity
summary: Två Home Assistant-instanser - "Krokviken" (10.20.20.20, HAOS 18.4) och "Skogsduvan" (stugan) - samt hur man administrations-API:t mot Supervisor nås via WebSocket.
tags: [home-assistant, ssh, hassio, supervisor, drift, projekt:simon]
created: 2026-11-25
updated: 2026-11-25
status: active
related: [notes/home-assistant-krokviken-ssh-via-terminal-ssh-addon, notes/skogsduvan-ssh-via-duckdns-fungerar-port-22-forwardad-2026-11-02, notes/ha-krokviken-robotdammsugare-startar-stoppar-pa-person-entiteter, notes/ha-tuya-tesvor-s6-anslutningen-fungerar-men-statusfalt-saknas]
---

## Sammanfattning

Det finns **två** Home Assistant-instanser som båda nås via SSH genom
Terminal & SSH-addonet (container `core-ssh`):

| Instans | Alias | Mål | Nyckel | Not |
|---|---|---|---|---|
| Krokviken (HAOS 18.4, aarch64) | `ssh ha`, `ssh krokviken`, `ssh homeassistant` | `root@10.20.20.20:22` | `~/.ssh/id_ed25519` (fungerar även med `~/.ssh/stuganha`) | LAN |
| Skogsduvan (stugan) | `ssh skogsduvan` | `root@stuganskogsduvan.duckdns.org:22` | `~/.ssh/stuganha` (RSA 4096) | Port 22 forwardad i stugans router sedan 2026-11-02 |

Host-entryn ligger i `~/.ssh/config`. Båda landar i **Terminal & SSH-addons
container** (hostname `core-ssh`), inte i HA-värden - `docker` och `ha` finns
där direkt, men `ha` kräver att man är i OS-skalet.

## Detaljer

### Krokviken (10.20.20.20)

- HAOS 18.4, Core 2026.11.3, Supervisor 2026.11.0, aarch64.
- Webben: `https://10.20.20.20:8123` (TLS direkt på 8123).
- Öppna portar: 8123 (HA), 4357 (ESPHome), 21064 (Nabu Casa Cloud).
- Terminal & SSH-addonet (core_ssh 10.6.0) är installerat och igång.

### Supervisor-API:t: `/api/hassio/*` ger 401

HA cores egna supervisor-endpoints (`/api/hassio/addons`,
`/addons/{slug}/config`, `/supervisor/info` ...) returnerar **401 även för
admin**. I HA core `dev` finns bara två whitelistar i
`homeassistant/components/hassio/http.py`:

- `PATHS_ADMIN` - bara backups, loggar, changelog, documentation
- `PATHS_NO_AUTH` - bara `addons/{slug}/(logo|icon)`

**Admins-API:t mot supervisor nås istället via WebSocket:**

```json
{"id": 1, "type": "supervisor/api", "endpoint": "/addons/core_ssh/info", "method": "GET"}
{"id": 1, "type": "supervisor/api", "endpoint": "/addons/core_ssh/options",
 "method": "POST", "data": {"options": {...}, "network": {"22/tcp": 22}}}
```

Kräver admin-token. Autentisering i websocket: `{"type":"auth","access_token":"<LLA-token>"}`.

### Aktivera SSH-porten på Terminal & SSH

Fellet "SSH port is disabled. Prevent start of SSH server." betyder att
addons nätverksmappning är `{"22/tcp": null}` - inget värdport publiceras.
Fix via supervisor-API:t:

```
POST /addons/core_ssh/options
{"options": {"authorized_keys": ["<pubkey>"], "password": "", "apks": [],
             "server": {"tcp_forwarding": false}},
 "network": {"22/tcp": 22}}
POST /addons/core_ssh/restart
```

Läste addon-loggen via `GET /api/hassio/addons/core_ssh/logs` (det är en
`PATHS_ADMIN`-sökväg och fungerar med Bearer-token).

### Vanliga fällor

- `authorized_keys` som lagras i addons options kan bli **trasig** av att
  nyckeln formatteras med mellanslag - kontrollera före `restart`.
- `hassio.addon_start` svarar 200 även om addonet inte startar - kontrollera
  alltid loggen (`/api/hassio/addons/{slug}/logs`) och porten.
- HA:s REST-svarar har tom `Server:`-header - core, inte nginx, ligger
  bakom 8123.
- `op item get <id> --reveal` fungerar i samma processsammanhang som
  1Password-appen (se minnet om op-cli-sessioner).

### Robotdammsugare (Tesvor S6)

Entiteten `vacuum.tesvor_s6` (Tuya cloud) har `supported_features: 13084`,
vilket ger `vacuum.start` och `vacuum.return_to_base` (bit 8192 resp. 16).

| Automation | id | Utlösare | Villkor | Åtgärd |
|---|---|---|---|---|
| Dammsug när vi lämnar huset | `1765542273027` | zone-leave för `person.simon_brundin` och `person.emma_wallin` (båda) | `not` (simon i zone.home) `and` `not` (emma i zone.home) | `vacuum.start` (utan target) |
| Dammsugare - ga tillbaka till station nar nagon kommer hem | `vacuum_return_when_someone_home` | state -> `home` för `person.simon_brundin` / `person.emma_wallin` | `not` (vacuum i `docked`/`charging`) | `vacuum.return_to_base` mot `vacuum.tesvor_s6` |

**Att skapa/uppdatera automationer utan UI:** `POST /api/config/automation/config/<id>`
(admin) med JSON-body `{"alias", "description", "triggers", "conditions", "actions", "mode"}`.
Samma id uppdaterar, `DELETE /api/config/automation/config/<id>` tar bort. Servern
skriver YAML till `/config/automations.yaml` och laddar om automationsdomänen.
Tricket `{"condition": "not", "conditions": [{...}]}` får INTE skrivas som
dubbla `condition`-nycklar i samma Python-dict - då försvinner `not` och
villkoret blir omvänt.

### Autentisering mot HA

Långlivstoken för HA (Core-API:t) finns i 1Password, item "Home Assistant"
(vault Private). Används bl.a. för `Authorization: Bearer <token>` mot
`https://10.20.20.20:8123/api/`. Websocket-login mot HA: skicka
`{"type":"auth","access_token":"<token>"}` efter `auth_required`.

## Relaterade

- [[notes/home-assistant-krokviken-ssh-via-terminal-ssh-addon]]
- [[notes/ha-krokviken-robotdammsugare-startar-stoppar-pa-person-entiteter]]
- [[notes/ha-tuya-tesvor-s6-anslutningen-fungerar-men-statusfalt-saknas]]
- [[notes/skogsduvan-ssh-via-duckdns-fungerar-port-22-forwardad-2026-11-02]]
- [[notes/skogsduvans-ha-temperaturlogik-ombyggd-fyra-doda-sensorer]]
