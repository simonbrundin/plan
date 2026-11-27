---
title: Skrivbordswidget som visar Gör-nu-målet
type: note
tags: [widget, quickshell, plan, gnistring, projekt:plan]
created: 2026-11-27
updated: 2026-11-27
status: active
related: [concepts/anvanda-agenter-i-plan, notes/neon-anslutningen-hanger-nar-ipv6-prioriteras, notes/platsfiltret-pa-now-ingen-plats-vald-visar-mal-utan-plats, notes/platsfiltret-pa-now-utan-plats-visar-mal-utan-plats]
---

# Skrivbordswidget för Gör-nu-målet (2026-11-27)

En liten ruta nederst i mitten av skärmen visar det mål som ligger högst upp
på /now. Klickar man på den öppnas målet i appen.

Rutan ligger i layer-shell-lagret **Overlay**, så den syns över alla program
och alla workspaces - även över fullskärmsfönster. Låsskärmen ligger också
i overlay-lagret men skapas senare, så den hamnar ovanför rutan.

Eftersom ytan går över hela skärmbredden sätts en **input-mask**, så att
bara själva rutan tar klick och resten av ytan går vidare till programmet
under.

## Rutan kan flyttas och döljas

Rutan är **utdragbar med musen**: tryck och dra. Klickar man utan att dra
öppnas målet i webbläsaren. Högerklick döljer rutan - praktiskt när man vill
se på en film. Samma sak går att göra med tangenten **SUPER SHIFT + .**
(`o.bind` i `~/.config/hypr/bindings.lua`).

Döljningen sker via en flagganfil (`~/.config/plan-goal-widget/hidden`) som
widgeten läser varje sekund. Filen läses med `Process` + `SplitParser`, inte
med `FileView` - `FileView` cachar innehållet och en ändring märks inte.

Var rutan ligger sparas i `position.txt` och läses tillbaka vid start.

### Varför bindningen ligger i `bindings.lua` och inte `bindings.conf`

Omarchy 2026 laddar bindningarna via Lua (`hyprland.lua` -> `hypr.bindings`).
Rader i `bindings.conf` registreras inte - `hyprctl binds` visar dem inte och
de fungerar inte. Bindningen läggs därför in med `o.bind(...)`.

### Muskoordinater i QML

`MouseArea.mouseX` är relativt till den item den ligger i. Eftersom rutan
flyttar sig under musen ger det en återkopplingsloop - rutan flyttade sig
två gånger så långt som musen rörde sig. Lösningen är att räkna på
**absoluta** koordinater: `box.x + event.x` (rutan ligger i fönstret, som
täcker hela skärmen, så fönstrets koordinater är skärmens).

`event.sceneX` fungerar inte - det är `undefined` i Quickshells
`QQuickMouseEvent` och ger "Cannot assign [undefined] to double".

## Platsfilter: mål utan plats syns också

`/now` filtrerade bort alla mål som inte var kopplade till den plats man
befinner sig på. Det är bakvänt - ett mål **utan plats** är inte bundet till
någon särskild plats och kan göras var som helst.

Fixen (PR #143) är `isDoableAtPlaceOrUnplaced` i `src/nuxt/app/utils/places.ts`:
flera platser har fortfarande OR-semantik, men ett mål utan platser räknas
som görbart överallt. Widgetens `nowlist.js` har samma regel, så appen och
widgeten visar exakt samma lista (14 mål, toppmål id 44).

Senare (2026-11-27) utökades regeln: är **ingen** plats vald ("Utan plats")
visas bara mål utan plats - de är de enda som är "görbara" när man inte vet
var man befinner sig. Platsfiltret i appen är dessutom påslaget som
standard, så appen och widgeten filtrerar likadant. Se
[[notes/platsfiltret-pa-now-utan-plats-visar-mal-utan-plats]].

## Vad siffran till höger betyder

Siffran är målets **delmål**: klara delmål / totalt antal delmål. Ett mål
utan delmål visar ingen siffra alls (t.ex. "Ladda batterier", som är ett
delmål till "Campervan", visar bara titeln).

Delmålen räknas ur `/api/v1/goals/graph` - den innehåller även mål som är
klara, så att framsteget kan räknas. Listan i övrigt byggs ur
`/goals/prioritized` och `/me/current-place`.

## Vakthund: widgeten ska alltid synnas

Det hände att widgeten slutade ritas (processen levde, men ytan syntes inte)
- troligen när Hyprland remappar sina lager. Därför finns en vakthund:

- `~/.config/plan-goal-widget/watchdog.sh` - fotograferar den gula punkten
  med `grim` och mäter hur stor del av området som är gul. Syns den inte
  tre gånger i rad startas widgeten om.
- `~/.config/systemd/user/plan-goal-widget-watchdog.timer` - kör vakthunden
  varannan minut.

Widgeten skriver sin geometri till `$XDG_RUNTIME_DIR/plan-goal-widget/geom.txt`
vid varje uppdatering - vakthunden vet därför var den ska titta.
(`grim` använder **logiska** pixlar, precis som geometrin.)

## Varför den ligger utanför repot

Widgeten är en Quickshell-instans, inte en del av Plan:

- `~/.config/quickshell/plan-goal/shell.qml` - gränssnittet
- `~/.config/quickshell/plan-goal/nowlist.js` - listlogiken (ren JS, testbar)
- `~/.config/systemd/user/plan-goal-widget.service` - startar widgeten
- `~/.config/plan-goal-widget/widget.env` - API-adress och nyckel (chmod 600)
- `~/.config/plan-goal-widget/toggle.sh` - visa/dölj rutan
Den körs som en egen Quickshell-process (`quickshell -p ~/.config/quickshell/plan-goal`)
så att den inte påverkas av `omarchy update` och inte hör ihop med Omarchy-skalet.

## Datan kommer från en egen API-instans

Datakällan är `http://127.0.0.1:4101/api/v1` - en egen Go-API-instans som
systemd kör mot Neon dev:

- `~/.config/systemd/user/plan-api-dev.service` - API:t (binär i ~/.local/bin/plan-api-dev)
- `~/.config/plan-goal-widget/api.env` - DATABASE_URL mot Neon (chmod 600)

Varför inte dev-API:t i Tilt-klustret (port 4100)? Då är widgeten beroende av
ett Tilt-fönster som är igång. Varför inte prod (plan.simonbrundin.com)?
**Prod-databasen (CNPG) är tom** - all användardata ligger i Neons dev-branch,
som dev-miljön kör mot. Widgeten visar alltså samma data som dev-appen.

## Auth: ny scope "readonly"

Widgeten har en egen API-nyckel med scope `readonly` - den får göra GET
men inte ändra något. Scopet lades till i API:t (Go och Rust) och i
databasschemat (`api_keys_valid_scope`). Nyckeln finns i `widget.env` och i
1Password (item "Plan - skrivbordswidget").

## Varför ingen ny endpoint i API:t

Först tänkte jag lägga `GET /api/v1/goals/now` i Go API:t, så att listan
byggs på servern. Men Plan API:t portas till Rust, och paritetstestet i
`src/rust` kräver att varje route i Go-tjänsten finns i Rust-tabellen - en
ny endpoint hade krävt en port till Rust också, och då hade samma logik
funnits i tre ställen (Go, Rust, Nuxt).

I stället läser widgeten de routes som redan finns (`/goals/prioritized` och
`/me/current-place`) och bygger listan i `nowlist.js`, som är en spegelbild
av gränssnittets `nowHierarchy.ts`, `goalFilters.ts` och `places.ts`.
Listan är testad mot TypeScript-implementationen: samma mål, samma ordning.

## Att ändra listlogiken

Ändras listan i appen ska `nowlist.js` ändras med. Testet som jämför
widgetens lista med appens TypeScript-logik:

```bash
curl -H "X-API-Key: $KEY" http://127.0.0.1:4101/api/v1/goals/prioritized -o /tmp/data/goals.json
curl -H "X-API-Key: $KEY" http://127.0.0.1:4101/api/v1/me/current-place -o /tmp/data/place.json
bun /tmp/jamfor-nowlist.ts   # kräver TS-filer utkopplade från src/nuxt/app/utils
```

## Att bygga om API-binären

`~/.local/bin/plan-api-dev` är en kopia av Go API:t och byggs om manuellt
när repot ändras:

```bash
cd <repo>/src && go build -o ~/.local/bin/plan-api-dev ./api/cmd/api
systemctl --user restart plan-api-dev
```

## Nyckeln

Nyckeln ligger i `~/.config/plan-goal-widget/widget.env` (chmod 600). I
1Password ligger den i itemet "Plan - skrivbordswidget" - lägg in den manuellt
om den saknas (`op item create --category=API Credential ...`).

## Relaterade

- [[notes/platsfiltret-pa-now-utan-plats-visar-mal-utan-plats]]
- [[notes/platsfiltret-pa-now-ingen-plats-vald-visar-mal-utan-plats]]
- [[notes/neon-anslutningen-hanger-nar-ipv6-prioriteras]]
- [[concepts/anvanda-agenter-i-plan]]
