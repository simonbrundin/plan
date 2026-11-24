---
title: MailMaster: kontodropdown + Outlook-inloggning via device code
type: note
tags: [mailmaster, feature, oauth, nuxt, go, projekt:simon]
---

## Vad som byggdes

Nuxt-UI:t har nu en dropdown med alla e-postkonton och en inloggningsknapp
som initierar en Microsoft-inloggning när kontot inte är inloggat.

## Varför device code och inte PKCE

Det gamla OAuth-flödet (`ai/auth.go`) var Authorization Code + PKCE med en
lokal callback-server som öppnade webbläsaren **på servern**. Det fungerar
inte från en webb-UI: browsaren kan ligga på en annan maskin, och
callback-URL:en `:8400/callback` kolliderar med Nuxt-devservern.

Device code-flödet (`src/go/oauth/oauth.go`) löser båda: servern frågar
Microsoft för en kod, användaren matar in koden i sin egen browser på
`https://login.microsoft.com/device`, och frontend pollar
`GET /api/v1/accounts/{konto}/auth/status` tills `status=complete`.

## Arkitektur

- `src/go/oauth/` - nytt paket. OAuth-koden flyttades hit från `ai/auth.go`
  (som var oanvänd kod - `ai`-paketet importerar inte längre oauth).
  Paketet har både device code och PKCE, plus token-cache på disk
  (`~/.cache/mailmaster/oauth-<hash>.json`, refresh-token stöds).
- `src/go/mcp/session.go` - **per-konto-sessioner**. Tidigare fanns en enda
  global IMAP-klient som stängdes när ett annat konto valdes. Nu finns en
  `sessionStore` med en session per kontonamn, så att växling är billig och
  konton inte blandas ihop vid samtidiga anrop.
- `src/go/mcp/auth.go` - HTTP-endpoints för kontolista, mappar och inloggning.
- `src/nuxt/app/composables/useAccounts.ts` - kontostate, inloggningsflöde,
  mapphantering per konto. `src/nuxt/app/utils/accounts.ts` har den rena
  logiken (testad med bun).

## API

- `GET  /api/v1/accounts` - konton + `auth_method`/`logged_in`/`login_needed`
- `GET  /api/v1/accounts/{konto}/folders` - mappar (kräver inloggning)
- `POST /api/v1/accounts/{konto}/auth/login` - starta device code, ger user_code
- `GET  /api/v1/accounts/{konto}/auth/status` - idle/waiting/complete/expired

## Buggar som upptäcktes under vägen

1. `imap.Client.Close()` panikade på en klient som aldrig anslutit
   (`c.conn.Close()` när conn var nil).
2. `go-imap:s Client.Close()` stänger *mboxen* (IMAP CLOSE-kommandot), inte
   TCP-anslutningen. Rätt sätt att stänga är `Terminate()`. Detta betydde att
   "stängd" session egentligen levde kvar.
3. `credentialFor` behöver OAuth-token för OAuth-konton i stället för
   lösenord ur env - annars gick `GetPassword()` fel.

## Kvar att lösa

Microsoft-kontot `michael.brundin@soulmate.se` kräver fortfarande
**admin consent** för appen `f377dda6-1879-413f-a89a-8a33e2f5f1c6`.
Device code-flödet fungerar tekniskt (kod skapas, polling fungerar), men
Microsoft svarar "Need admin approval" i steget där användaren ska godkänna
behörigheterna. Två vägar: (a) en admin ger admin consent, (b) byt client_id
till en etablerad public client (t.ex. Thunderbirds
`9e5f94bc-e8a4-4e73-b8be-63364c29d753` eller Outlooks
`d3590ed6-52b3-4102-aeff-aad2292ab01c`).

## Verifiering

- Go: `go test ./...` - 23 tester i oauth + mcp gröna.
- Nuxt: `bun test` - 23 tester gröna.
- Browser (Playwright mot containern): dropdown fylls med båda kontona,
  "Inloggad"-badge för one.com, "Logga in"-knapp för Outlook-kontot, kod visas,
  polling svarar `waiting`, mejl hämtas (2 mejl i INBOX.Simon).
