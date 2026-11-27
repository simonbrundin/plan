---
title: Skrivbordswidget som visar Gör-nu-målet
type: note
tags: [widget, quickshell, plan, gnistring, projekt:plan]
---

# Skrivbordswidget för Gör-nu-målet (2026-11-27)

En liten ruta nederst i mitten av skärmen visar det mål som ligger högst upp
på /now. Klickar man på den öppnas målet i appen. Rutan ligger i
layer-shell-lagret "Top", så den ligger över andra program men under
notifikationer.

## Varför den ligger utanför repot

Widgeten är en Quickshell-instans, inte en del av Plan:

- `~/.config/quickshell/plan-goal/shell.qml` - gränssnittet
- `~/.config/quickshell/plan-goal/nowlist.js` - listlogiken (ren JS, testbar)
- `~/.config/systemd/user/plan-goal-widget.service` - startar widgeten
- `~/.config/plan-goal-widget/widget.env` - API-adress och nyckel (chmod 600)

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
