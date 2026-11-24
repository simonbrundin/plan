---
title: Mailagentens MCP-verktygen saknade konto och hade fel schema
type: note
tags: [bug, mcp, mail, go, projekt:pappa]
created: 2026-11-16
updated: 2026-11-17
status: active
related: [notes/mailagenten-flyttad-fran-mailmaster-repot-in-i-pappa, notes/mailagentens-kontodropdown-renderades-aldrig-nuxt-prefixade-komponenter, notes/mailmaster-svarmejl-fick-fel-brodtext-tre-buggar-i-svarttextskedjan]
---

2026-11-16: Alla sju MCP-verktygen i pappa/src/go/mcp gick mot det FÖRSTA kontoet - det fanns ingen möjlighet för en agent att välja konto, trots att config.yaml har två (Michael Brundins / one.com och Soulmate / Office365).

## Två buggar

1. Ingen kontoparameter. Varje verktyg anropade ensureConnection() (= första kontot). Nu har alla params-strukturer ett Account-fält (omitempty) och verktygen anropar ensureConnectionFor(args.Account). Tomt = första kontot, så bakåtkompatibiliteten hålls.

2. Fel JSON-schema. folder och limit stod som required i schemat, trots att koden har defaultvärden (resolveFolder, limit=10). jsonschema-go infererar required från omitempty-taggen - utan taggen blir fältet required. Följden: en agent som anropade list_emails med bara {"account": "Soulmate"} fick validatoringsfelet 'required: missing properties: ["folder" "limit"]'. Detta var redan trasigt FÖRE ändringen - den gamla versionen krävde också folder och limit.

## Sätt att hitta required-fält

Kör ett riktigt tools/list-anrop mot servern och titta på
inputSchema.required. Eller: jsonschema.ForType(reflect.TypeOf(params),
&jsonschema.ForOptions{}) - det är exakt samma inferens som MCP-servern
gör, så testet ser vad agenten ser.

## Verktygsbeskrivningarna

toolDefinitions() bygger beskrivningarna och skriver in kontonamnen
('Accounts: Michael Brundins, Soulmate') så en agent ser vilka konton
som finns utan att behöva lista dem först. list_accounts returnerar
kontona som JSON med inloggningsstatus.

## Relaterat

[[notes/mailagentens-kontodropdown-renderades-aldrig-nuxt-prefixade-komponenter]]

## Relaterade

- [[notes/mailmaster-svarmejl-fick-fel-brodtext-tre-buggar-i-svarttextskedjan]]
- [[notes/mailagentens-kontodropdown-renderades-aldrig-nuxt-prefixade-komponenter]]
- [[notes/mailagenten-flyttad-fran-mailmaster-repot-in-i-pappa]]
